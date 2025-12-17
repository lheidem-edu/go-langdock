package langdock

import (
	"bytes"
	"context"
	"encoding/json"
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

type UpdateKnowledgeFileRequest struct {
	FolderID     uuid.UUID
	AttachmentID uuid.UUID
	FileName     string
	Content      io.Reader
}

type UpdateKnowledgeFileResponse struct {
	Status string        `json:"status"`
	Result KnowledgeFile `json:"result"`
}

func (k *KnowledgeService) UpdateKnowledgeFile(ctx context.Context, params *UpdateKnowledgeFileRequest) (*UpdateKnowledgeFileResponse, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", params.FileName)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, params.Content); err != nil {
		return nil, err
	}

	if err := writer.WriteField("attachmentId", params.AttachmentID.String()); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/knowledge/%s", params.FolderID)
	req, err := k.Client.NewRequest(ctx, http.MethodPatch, path, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	var resp UpdateKnowledgeFileResponse
	if err := k.Client.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

type DeleteKnowledgeFileRequest struct {
	FolderID     uuid.UUID
	AttachmentID uuid.UUID
}

func (k *KnowledgeService) DeleteKnowledgeFile(ctx context.Context, params *DeleteKnowledgeFileRequest) error {
	path := fmt.Sprintf("/knowledge/%s/%s", params.FolderID, params.AttachmentID)
	req, err := k.Client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return k.Client.Do(req, nil)
}

type SearchResult struct {
	Text       string    `json:"text"`
	Similarity float64   `json:"similarity"`
	Subsource  string    `json:"subsource"`
	Subname    string    `json:"subname"`
	ID         uuid.UUID `json:"id"`
	URL        string    `json:"url"`
	Index      int       `json:"index"`
}

type SearchKnowledgeFilesRequest struct {
	Query string `json:"query"`
}

type SearchKnowledgeFilesResponse struct {
	Status string         `json:"status"`
	Result []SearchResult `json:"result"`
}

func (k *KnowledgeService) SearchKnowledgeFiles(ctx context.Context, params *SearchKnowledgeFilesRequest) (*SearchKnowledgeFilesResponse, error) {
	path := "/knowledge/search"
	req, err := k.Client.NewRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}

	var encoded bytes.Buffer
	if err := json.NewEncoder(&encoded).Encode(params); err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(&encoded)
	req.Header.Set("Content-Type", "application/json")

	var resp SearchKnowledgeFilesResponse
	if err := k.Client.Do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
