// Package main provides a simple HTTP server to serve videos from an Azure Blob Storage.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

func getBlobToStream(accountName, accountKey, containerName, blobName string) (io.ReadCloser, error) {
	// Create Shared Key Credential
	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, err
	}

	// Create a request pipeline to process HTTP(S) requests and responses
	pipeline := azblob.NewPipeline(cred, azblob.PipelineOptions{})
	URL, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))

	// Create a new container URL
	containerURL := azblob.NewContainerURL(*URL, pipeline)

	// Get a blob URL; this URL is used to access and manipulate the blob
	blobURL := containerURL.NewBlobURL(blobName)

	// Download the blob's content
	resp, err := blobURL.Download(context.Background(), 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return nil, err
	}

	return resp.Body(azblob.RetryReaderOptions{MaxRetryRequests: 20}), nil
}

// videoHandler serves a video from Azure Blob Storage when accessed via an HTTP GET request.
// It responds with "Method not allowed" for non-GET requests.
// It uses environment variables STORAGE_ACCOUNT_NAME and STORAGE_ACCESS_KEY to authenticate with Azure Blob Storage.
func videoHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the HTTP method is GET
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		w.Write([]byte("Method not allowed"))
		return
	}

	log.Println("Checking environment variables...")

	// Retrieve environment variables
	accountName, ok1 := os.LookupEnv("STORAGE_ACCOUNT_NAME")
	accountKey, ok2 := os.LookupEnv("STORAGE_ACCESS_KEY")

	// Ensure environment variables are set
	if !ok1 || !ok2 {
		http.Error(w, "Environment variables not set", http.StatusInternalServerError)
		log.Println("Environment variables not set!")
		return
	}

	containerName := "videos"

	// Get the blob name from the 'path' query parameter
	blobName := r.URL.Query().Get("path")
	if blobName == "" {
		http.Error(w, "path query parameter is missing", http.StatusBadRequest)
		return
	}

	log.Printf("Retrieving blob: %s from container: %s", blobName, containerName)

	reader, err := getBlobToStream(accountName, accountKey, containerName, blobName)

	// Get the video blob stream
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving video: %v", err), http.StatusInternalServerError)
		log.Printf("Error retrieving blob: %v", err)
		return
	}

	log.Println("Blob retrieved successfully.")

	defer reader.Close()

	// Set the response headers and write the video content
	w.Header().Set("Content-Type", "video/mp4")
	io.Copy(w, reader)
}

// main initializes an HTTP server that listens on port 8080 and serves videos from Azure Blob Storage.
func main() {
	http.HandleFunc("/video", videoHandler)
	http.ListenAndServe(":80", nil)
}
