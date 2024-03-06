package formdata

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/rus-sharafiev/go-rest/common"
	"github.com/rus-sharafiev/go-rest/common/auth"
	"github.com/rus-sharafiev/go-rest/common/exception"
)

type uploadErrors struct {
	UploadErrors []uploadError `json:"uploadErrors"`
}

type uploadError struct {
	Filename string `json:"filename"`
	Error    string `json:"error"`
}

// Form Data interceptor for custom JSON to Form Data converter.
// The source code of the converter can be found at
// https://github.com/rus-sharafiev/fetch-api/blob/master/src/fetch-api.ts#L180
func Interceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check for whitelist and blacklist
		if wl := common.Config.FormdataWhitelist; wl != nil && !inList(*wl, r) {
			next.ServeHTTP(w, r)
			return
		} else if bl := common.Config.FormdataBlacklist; wl == nil && bl != nil && inList(*bl, r) {
			next.ServeHTTP(w, r)
			return
		}

		// Check whether request contains multipart/form-data
		if mr, err := r.MultipartReader(); err == nil {

			// Use user folder if user is authorized or return 401
			userDir := ""
			if userID, _ := auth.Headers(r); len(userID) != 0 {
				userDir = userID
			} else {
				exception.UnauthorizedError(w, fmt.Errorf("only authorized users can save files"))
				return
			}

			// Check if subdirectory exists
			fullPath := path.Join(*common.Config.UploadDir, userDir)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				if err := os.Mkdir(fullPath, 0755); err != nil {
					exception.InternalServerError(w, err)
					return
				}
			}

			// Read form
			form, err := mr.ReadForm(32 << 20)
			if err != nil {
				exception.InternalServerError(w, err)
				return
			}

			resultChan := make(chan []string)
			errChan := make(chan uploadError)
			fileList := []string{}

			// Iterate over the range of multipart file fields
			for name, values := range form.File {
				for index, fileHeader := range values {
					fileList = append(fileList, name+"-"+strconv.Itoa(index))
					fileHeader := fileHeader
					name := name

					go func() {

						// Get file
						file, err := fileHeader.Open()
						if err != nil {
							errChan <- uploadError{Filename: fileHeader.Filename, Error: err.Error()}
							return
						}

						// Generate unique file name
						id, err := uuid.NewRandom()
						if err != nil {
							errChan <- uploadError{Filename: fileHeader.Filename, Error: err.Error()}
							return
						}

						fileName := path.Join(fullPath, id.String()+filepath.Ext(fileHeader.Filename))

						// Write
						outFile, err := os.Create(fileName)
						if err != nil {
							errChan <- uploadError{Filename: fileHeader.Filename, Error: err.Error()}
							return
						}
						defer outFile.Close()

						_, err = io.Copy(outFile, file)
						if err != nil {
							errChan <- uploadError{Filename: fileHeader.Filename, Error: err.Error()}
							return
						}

						resultChan <- []string{name, strings.Replace(fileName, fullPath, *common.Config.UploadDir, 1)}
					}()
				}

			}

			// Create uploaded files map and errors slice
			filesUrlMap := make(map[string]interface{})
			errorSlice := []uploadError{}

			for range fileList {
				select {
				case result := <-resultChan:
					name := result[0]
					fileName := result[1]

					// Add uploaded file url to map as string or append to existing slice
					if filesUrlMap[name] == nil {

						filesUrlMap[name] = fileName

					} else {

						if existingStr, ok := filesUrlMap[name].(string); ok {
							filesUrlMap[name] = []string{existingStr, fileName}
						}

						if existingSlice, ok := filesUrlMap[name].([]string); ok {
							filesUrlMap[name] = append(existingSlice, fileName)
						}
					}

				case err := <-errChan:
					errorSlice = append(errorSlice, err)
				}
			}

			var resultJson string

			// Convert map with files urls to JSON object
			if len(filesUrlMap) != 0 {
				fileList, err := json.Marshal(filesUrlMap)
				if err != nil {
					exception.InternalServerError(w, err)
					return
				}
				resultJson = string(fileList)
			}

			// Add errors to result
			if len(errorSlice) != 0 {
				uploadErrors, err := json.Marshal(uploadErrors{UploadErrors: errorSlice})
				if err != nil {
					exception.InternalServerError(w, err)
					return
				}
				if len(resultJson) != 0 {
					resultJson = strings.TrimSuffix(resultJson, "}") + "," + strings.TrimPrefix(string(uploadErrors), "{")
				} else {
					resultJson = string(uploadErrors)
				}
			}

			// Add `serialized-json` field value and JSON object with files urls to result JSON
			if jsonField := form.Value["serialized-json"]; len(jsonField) != 0 {
				if len(resultJson) != 0 {
					resultJson = strings.TrimSuffix(jsonField[0], "}") + "," + strings.TrimPrefix(resultJson, "{")
				} else {
					resultJson = jsonField[0]
				}
			}

			// Add errors if exists

			// Write result JSON string to request body
			w.Header().Set("Content-Type", "application/json")
			r.Body = io.NopCloser(strings.NewReader(resultJson))
		}

		next.ServeHTTP(w, r)
	})
}

// Check if one of the list strings matches the request URL path
func inList(list []string, r *http.Request) bool {
	path := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	return slices.ContainsFunc(list, func(s string) bool {
		return slices.Contains(path, s)
	})
}
