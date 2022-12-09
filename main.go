package main

import (
	"archive/zip"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	imageBucket  = os.Getenv("IMAGE_BUCKET")
	uploadBucket = os.Getenv("UPLOAD_BUCKET")

	s3Client      *s3.Client
	presignClient *s3.PresignClient
	uploader      *manager.Uploader

	//go:embed index.tmpl
	htmlTemplate string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	s3Client = s3.NewFromConfig(cfg)
	presignClient = s3.NewPresignClient(s3Client)
	uploader = manager.NewUploader(s3Client)
}

func main() {
	ctx := context.Background()
	if err := os.Mkdir("downloads", os.ModePerm); err != nil {
		fmt.Println(err)
	}
	fmt.Println("Downloading and zipping files...")
	zipFileName, err := zipFiles(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Wrote zip file to " + zipFileName)
	if err := uploadZip(ctx, zipFileName); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Uploaded zip file to " + uploadBucket + "/" + zipFileName)
	presigned, err := presignURL(ctx, zipFileName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Pre-signed URL: " + presigned)
}

func zipFiles(ctx context.Context) (string, error) {
	names := getFileNames()
	// バッチだとos.CreateTemp()を使う
	zipArchive, err := os.Create("downloads/out_" + strconv.FormatInt(time.Now().Unix(), 10) + ".zip")
	if err != nil {
		return "", err
	}
	defer func() { _ = zipArchive.Close() }()
	writer := zip.NewWriter(zipArchive)
	defer func() { _ = writer.Close() }()

	if err := generateAndWriteHTML(writer, names); err != nil {
		return "", err
	}

	// 画像をDL＆zipに書き込む
	for _, name := range names {
		// Note: wはio.Writerなのでmanager.Downloaderには使えない（io.WriterAtが必要）
		// なので普通にs3.GetObjectのリクエストでダウンロードする
		w, err := writer.CreateHeader(&zip.FileHeader{
			Name:     name,
			Modified: time.Now(),
		})
		if err != nil {
			fmt.Println(err)
			continue
		}
		if err := downloadImage(ctx, name, w); err != nil {
			fmt.Println(err)
		}
	}
	return zipArchive.Name(), nil
}

func generateAndWriteHTML(writer *zip.Writer, imgNames []string) error {
	tpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		return err
	}
	w, err := writer.CreateHeader(&zip.FileHeader{
		Name:     "index.html",
		Modified: time.Now(),
	})
	if err != nil {
		return err
	}
	if err := tpl.Execute(w, imgNames); err != nil {
		return err
	}
	return nil
}

func downloadImage(ctx context.Context, name string, w io.Writer) error {
	out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(imageBucket),
		Key:    aws.String(name),
	})
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, out.Body); err != nil {
		return err
	}
	return nil
}

func uploadZip(ctx context.Context, fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(uploadBucket),
		Key:    aws.String(fileName),
		Body:   f,
	})
	return err
}

func presignURL(ctx context.Context, fileName string) (string, error) {
	out, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(uploadBucket),
		Key:    aws.String(fileName),
	}, s3.WithPresignExpires(time.Minute*5))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func getFileNames() []string {
	files, err := os.ReadDir("images")
	if err != nil {
		log.Fatal(err)
	}
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Name()
	}
	return names
}
