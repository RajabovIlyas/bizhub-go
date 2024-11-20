package helpers

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path"
	"strings"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var ImageExtensions = []string{"jpg", "jpeg", "png", "webp"}

func DeleteImages(files []string) {
	for _, f := range files {
		DeleteImageFile(f)
	}
}
func DeleteImageFile(file string) {
	if len(file) == 0 {
		log.Printf("\nfile name length = 0.\n")
		return
	}
	if strings.Contains(file, "default_images") {
		log.Printf("\nnot deleting default image: %v\n", file)
		return // default_images directory images not to be deleted!
	}
	imagePath := path.Join(config.RootPath, "public", file)
	fmt.Printf("\ndeleting image path: %v\n", imagePath)
	e := os.Remove(imagePath)
	if e != nil {
		log.Printf("\nCouldn't delete image file: %v\nError: %v\n", file, e) // os.Exit() etmek gerek dalmika diyyan?!!
	}
}
func SaveFileheader(c *fiber.Ctx, file *multipart.FileHeader, folder string) (string, error) {
	// fmt.Printf("\nFile Header: %v", file.Header)
	// fmt.Printf("\nFile Size: %v\n", file.Size)
	// fmt.Printf("\nFile Name: %v\n", file.Filename)
	s := strings.Split(file.Filename, ".")
	extension := strings.ToLower(s[len(s)-1])
	fmt.Printf("\nextension is: %v\n", extension)

	ext := strings.Split(file.Header["Content-Type"][0], "/")[1]

	fmt.Printf("\nextension from content-type: %v\n", ext)
	// if !SliceContains(ImageExtensions, ext) {
	// 	return "", fmt.Errorf("File: %v - is not a valid image file.", file.Filename)
	// }

	// if !SliceContains(ImageExtensions, extension) {
	// 	return "", fmt.Errorf("File: %v - is not a valid image file.", file.Filename)
	// }
	rootPath := config.RootPath
	newFileUuid := uuid.New()
	splitUuid := strings.Split(newFileUuid.String(), "-")
	newFileName := strings.Join(splitUuid, "") + "." + extension
	newFilePath := path.Join("images", folder, newFileName)
	imagePath := path.Join(rootPath, "public", newFilePath)
	err := c.SaveFile(file, imagePath)
	if err != nil {
		return "", err
	}
	return newFilePath, nil

}
func SaveImageFile(c *fiber.Ctx, imageField string, folder string) (string, error) {
	file, err := c.FormFile(imageField)
	if err != nil {
		return "", err
	}
	// fmt.Printf("\nFile Header: %v", file.Header)
	// fmt.Printf("\nFile Size: %v\n", file.Size)
	// fmt.Printf("\nFile Name: %v\n", file.Filename)
	s := strings.Split(file.Filename, ".")
	extension := strings.ToLower(s[len(s)-1])
	fmt.Printf("\nextension is: %v\n", extension)
	if !SliceContains(ImageExtensions, extension) {
		return "", fmt.Errorf("File: %v - is not a valid image file.", file.Filename)
	}
	ext := strings.Split(file.Header["Content-Type"][0], "/")[1]
	fmt.Printf("\nextension from content-type: %v\n", ext)
	// if !SliceContains(ImageExtensions, ext) {
	// 	return "", fmt.Errorf("File: %v - is not a valid image file.", file.Filename)
	// }

	rootPath := config.RootPath
	newFileUuid := uuid.New()
	splitUuid := strings.Split(newFileUuid.String(), "-")
	newFileName := strings.Join(splitUuid, "") + "." + extension
	newFilePath := path.Join("images", folder, newFileName)
	imagePath := path.Join(rootPath, "public", newFilePath)
	err = c.SaveFile(file, imagePath)
	if err != nil {
		return "", err
	}
	return newFilePath, nil
}
func CopyFile(src, dst string) (int64, error) {
	fileInfo, err := os.Stat(src)
	if err != nil {
		return 0, err
	}
	if !fileInfo.Mode().IsRegular() {
		return 0, fmt.Errorf("%v - is not a regular file.", src)
	}
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()
	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	n, err := io.Copy(destination, source)
	return n, err
}
