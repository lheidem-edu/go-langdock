package langdock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
)

type KnowledgeService struct {
	Client *Client
}

func NewKnowledgeService(client *Client) *KnowledgeService {
	return &KnowledgeService{
		Client: client,
	}
}

type KnowledgeFile struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	MimeType   string    `json:"mimeType"`
	CreatedAt  string    `json:"createdAt"`
	UpdatedAt  string    `json:"updatedAt"`
	URL        *string   `json:"url"`
	Path       *string   `json:"path,omitempty"`
	SyncStatus *string   `json:"syncStatus,omitempty"`
	PageCount  *int      `json:"pageCount,omitempty"`
	Summary    *string   `json:"summary,omitempty"`
}

type ListKnowledgeFilesRequest struct {
	FolderID uuid.UUID
}

type ListKnowledgeFilesResponse struct {
	Status string          `json:"status"`
	Result []KnowledgeFile `json:"result"`
}

func (k *KnowledgeService) ListKnowledgeFiles(ctx context.Context, params *ListKnowledgeFilesRequest) (*ListKnowledgeFilesResponse, error) {
	path := fmt.Sprintf("/knowledge/%s/list", params.FolderID)

	req, err := k.Client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var listing ListKnowledgeFilesResponse
	if err := k.Client.Do(req, &listing); err != nil {
		return nil, err
	}

	return &listing, nil
}

type UploadKnowledgeFileRequest struct {
	FolderID uuid.UUID
	FileName string
	Content  io.Reader
}

type UploadKnowledgeFileResponse struct {
	Status string        `json:"status"`
	Result KnowledgeFile `json:"result"`
}

func (k *KnowledgeService) UploadKnowledgeFile(ctx context.Context, params *UploadKnowledgeFileRequest) (*UploadKnowledgeFileResponse, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", params.FileName)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, params.Content); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/knowledge/%s", params.FolderID)
	req, err := k.Client.NewRequest(ctx, http.MethodPost, path, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	var resp UploadKnowledgeFileResponse
	if err := k.Client.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
