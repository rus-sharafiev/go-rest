package images

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rus-sharafiev/go-rest/common/auth"
	"github.com/rus-sharafiev/go-rest/common/exception"
	"golang.org/x/image/draw"
)

// SERVE UPLOAD -------------------------------------------------------------------
func (c *Controller) serve(w http.ResponseWriter, r *http.Request) {
	userId, role := auth.Headers(r)
	if len(userId) == 0 {
		exception.Unauthorized(w)
		return
	}

	uploadDir := filepath.Join(c.UploadDir, userId)
	if role == "ADMIN" {
		uploadDir = c.UploadDir
	}

	w.Header().Add("Cache-Control", "private, max-age=31536000, immutable")
	http.StripPrefix("/images/", http.FileServer(http.Dir(uploadDir))).ServeHTTP(w, r)
}

// HANDLE UPLOAD ------------------------------------------------------------------
func (c *Controller) handle(w http.ResponseWriter, r *http.Request) {
	const basePath = "/images"

	// Check whether request contains multipart/form-data
	if mr, err := r.MultipartReader(); err == nil {

		// Use user folder if user is authorized
		subFolder := ""
		if userID, _ := auth.Headers(r); len(userID) != 0 {
			subFolder = userID
		}

		// Check existence or create folders
		sizes := [5]string{Original, Small, Medium, Large, ExtraLarge}
		for _, size := range sizes {
			fullPath := path.Join(c.UploadDir, subFolder, size)

			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				if err := os.Mkdir(fullPath, 0755); err != nil {
					exception.InternalServerError(w, err)
					return
				}
			}
		}

		// Read form
		form, err := mr.ReadForm(32 << 20)
		if err != nil {
			exception.InternalServerError(w, err)
			return
		}

		imagesChan := make(chan GalleryImage)
		errChan := make(chan error)
		fileList := []string{}

		// Iterate over the range of multipart file fields
		for name, values := range form.File {
			for i, fileHeader := range values {
				fileList = append(fileList, name+"-"+strconv.Itoa(i))
				fileHeader := fileHeader

				go func() {
					// Get file
					file, err := fileHeader.Open()
					if err != nil {
						errChan <- err
						return
					}

					// Decode images
					var img image.Image
					switch contentType := fileHeader.Header.Get("Content-Type"); contentType {
					case "image/png":
						img, _ = png.Decode(file)
					case "image/jpeg":
						img, _ = jpeg.Decode(file)
					case "image/heic":
						errChan <- fmt.Errorf("error loading file %s: heic is not supported yet", fileHeader.Filename)
						return
					default:
						errChan <- fmt.Errorf("error loading file %s: file type is not supported. png, jpeg, heic are allowed", fileHeader.Filename)
						return
					}

					imgWidth := img.Bounds().Bounds().Max.X
					imgHeight := img.Bounds().Bounds().Max.Y
					aspectRatio := float32(imgWidth) / float32(imgHeight)

					// Create file name
					id, err := uuid.NewRandom()
					if err != nil {
						errChan <- err
						return
					}

					fileName := id.String() + ".jpg" // filepath.Ext(fileHeader.Filename)
					originalOutputLocation := ""

					// Create sizes
					imageSizesChan := make(chan []string)
					errorSizesChan := make(chan error)

					for _, size := range sizes {
						size := size

						go func() {

							// Create output file
							outputLocation := path.Join(path.Join(c.UploadDir, subFolder, size), fileName)
							outFile, err := os.Create(outputLocation)
							if err != nil {
								errorSizesChan <- err
								return
							}
							defer outFile.Close()

							// Write the original as PNG
							if size == Original {

								if err := jpeg.Encode(outFile, img, &jpeg.Options{Quality: 100}); err != nil {
									errorSizesChan <- err
									return
								}

							} else {

								// Get target size
								intSize, err := strconv.Atoi(size)
								if err != nil {
									errorSizesChan <- err
									return
								}

								if intSize >= max(imgWidth, imgHeight) {

									// Return original if smaller than target size
									if err := jpeg.Encode(outFile, img, nil); err != nil {
										errorSizesChan <- err
										return
									}

								} else {

									// Create new image with long side target size
									var dst *image.RGBA
									if aspectRatio >= 1 {
										dst = image.NewRGBA(image.Rect(0, 0, intSize, int(float32(intSize)/aspectRatio)))
									} else {
										dst = image.NewRGBA(image.Rect(0, 0, int(float32(intSize)*aspectRatio), intSize))
									}

									// Resize
									draw.BiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)

									// Encode to jpeg
									if err := jpeg.Encode(outFile, dst, nil); err != nil {
										errorSizesChan <- err
										return
									}
								}
							}
							imageSizesChan <- []string{size, strings.Replace(outputLocation, c.UploadDir, basePath, 1)}
						}()
					}

					imageSizes := make(map[string]string)
					var errorSizes error

					for range sizes {
						select {
						case sizes := <-imageSizesChan:
							if sizes[0] == Original {
								originalOutputLocation = sizes[1]
							} else {
								imageSizes[sizes[0]] = sizes[1]
							}
						case err := <-errorSizesChan:
							errorSizes = err
						}
					}

					if errorSizes != nil {
						errChan <- errorSizes
						return
					}

					// Create Gallery Image object
					galleryImage := GalleryImage{
						File:         &fileName,
						Name:         &fileHeader.Filename,
						Url:          &originalOutputLocation,
						OriginalSize: &fileHeader.Size,
						Width:        &imgWidth,
						Height:       &imgHeight,
						Sizes:        &imageSizes,
						CreatedAt:    time.Now(),
					}

					imagesChan <- galleryImage
				}()
			}

		}

		// Get results and errors from chain
		var galleryImageSlice []GalleryImage
		var errorsSlice []error

		for range fileList {
			select {
			case galleryImage := <-imagesChan:
				galleryImageSlice = append(galleryImageSlice, galleryImage)
			case err := <-errChan:
				errorsSlice = append(errorsSlice, err)
			}
		}

		for _, err := range errorsSlice {
			fmt.Println(err)
		}

		var resultJson string
		// Convert map with files urls to JSON object
		if len(galleryImageSlice) != 0 {
			filesList, err := json.Marshal(galleryImageSlice)
			if err != nil {
				exception.InternalServerError(w, err)
				return
			}
			resultJson = string(filesList)
		}

		// Write result JSON string to request body
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, resultJson)
	} else {
		exception.BadRequest(w)
	}
}
