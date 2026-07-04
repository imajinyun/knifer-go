package ssh

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeProvider struct {
	runRequests      []CommandRequest
	listRequests     []ListRequest
	downloadRequests []DownloadRequest
	uploadRequests   []UploadRequest
	runResponse      CommandResponse
	listResponse     ListResponse
	downloadResponse DownloadResponse
	uploadResponse   UploadResponse
	err              error
}

func (p *fakeProvider) Run(ctx context.Context, request CommandRequest) (CommandResponse, error) {
	select {
	case <-ctx.Done():
		return CommandResponse{}, ctx.Err()
	default:
	}
	p.runRequests = append(p.runRequests, request)
	return p.runResponse, p.err
}

func (p *fakeProvider) List(ctx context.Context, request ListRequest) (ListResponse, error) {
	select {
	case <-ctx.Done():
		return ListResponse{}, ctx.Err()
	default:
	}
	p.listRequests = append(p.listRequests, request)
	return p.listResponse, p.err
}

func (p *fakeProvider) Download(ctx context.Context, request DownloadRequest) (DownloadResponse, error) {
	select {
	case <-ctx.Done():
		return DownloadResponse{}, ctx.Err()
	default:
	}
	p.downloadRequests = append(p.downloadRequests, request)
	return p.downloadResponse, p.err
}

func (p *fakeProvider) Upload(ctx context.Context, request UploadRequest) (UploadResponse, error) {
	select {
	case <-ctx.Done():
		return UploadResponse{}, ctx.Err()
	default:
	}
	p.uploadRequests = append(p.uploadRequests, request)
	return p.uploadResponse, p.err
}

func TestClientRunUsesProviderAndClones(t *testing.T) {
	provider := &fakeProvider{runResponse: CommandResponse{ExitCode: 0, Stdout: []byte("ok")}}
	client := New(WithProvider(provider))
	request := CommandRequest{Command: "printf", Args: []string{"ok"}, Metadata: map[string]string{"trace": "one"}}

	response, err := client.Run(context.Background(), request)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	request.Args[0] = "changed"
	request.Metadata["trace"] = "changed"
	provider.runResponse.Stdout[0] = 'O'
	if provider.runRequests[0].Args[0] != "ok" || provider.runRequests[0].Metadata["trace"] != "one" {
		t.Fatalf("provider request was not cloned: %+v", provider.runRequests[0])
	}
	if string(response.Stdout) != "ok" {
		t.Fatalf("response was not cloned: %+v", response)
	}
}

func TestClientListUsesProviderAndClones(t *testing.T) {
	provider := &fakeProvider{listResponse: ListResponse{Entries: []Entry{{Name: "file.txt", Path: "/file.txt", Type: EntryTypeFile}}}}
	client := New(WithProvider(provider))
	request := ListRequest{RemoteDir: "/pub", Metadata: map[string]string{"trace": "one"}}

	response, err := client.List(context.Background(), request)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	request.Metadata["trace"] = "changed"
	provider.listResponse.Entries[0].Name = "changed"
	if provider.listRequests[0].Metadata["trace"] != "one" {
		t.Fatalf("provider request was not cloned: %+v", provider.listRequests[0])
	}
	if response.Entries[0].Name != "file.txt" {
		t.Fatalf("response was not cloned: %+v", response)
	}
}

func TestNilProviderOptionDoesNotOverwriteConfiguredProvider(t *testing.T) {
	provider := &fakeProvider{runResponse: CommandResponse{ExitCode: 0}}
	client := New(WithProvider(provider), WithProvider(nil))
	if _, err := client.Run(context.Background(), CommandRequest{Command: "true"}); err != nil {
		t.Fatalf("Run with nil overwrite option error = %v", err)
	}
	if len(provider.runRequests) != 1 {
		t.Fatalf("provider calls = %d, want 1", len(provider.runRequests))
	}
}

func TestClientRequiresProvider(t *testing.T) {
	client := New()
	if _, err := client.Run(context.Background(), CommandRequest{Command: "true"}); !errors.Is(err, ErrMissingProvider) {
		t.Fatalf("Run error = %v, want ErrMissingProvider", err)
	}
	if _, err := client.List(context.Background(), ListRequest{RemoteDir: "/"}); !errors.Is(err, ErrMissingProvider) {
		t.Fatalf("List error = %v, want ErrMissingProvider", err)
	}
	if _, err := client.Download(context.Background(), DownloadRequest{RemotePath: "/file.txt"}); !errors.Is(err, ErrMissingProvider) {
		t.Fatalf("Download error = %v, want ErrMissingProvider", err)
	}
	if _, err := client.Upload(context.Background(), UploadRequest{RemotePath: "/file.txt"}); !errors.Is(err, ErrMissingProvider) {
		t.Fatalf("Upload error = %v, want ErrMissingProvider", err)
	}
}

func TestClientValidatesBeforeProvider(t *testing.T) {
	provider := &fakeProvider{}
	client := New(WithProvider(provider))
	_, err := client.Run(context.Background(), CommandRequest{})
	if !errors.Is(err, ErrInvalidCommandRequest) {
		t.Fatalf("Run error = %v, want ErrInvalidCommandRequest", err)
	}
	if len(provider.runRequests) != 0 {
		t.Fatalf("provider was called for invalid request")
	}
}

func TestClientPropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := New(WithProvider(&fakeProvider{}))
	_, err := client.Download(ctx, DownloadRequest{RemotePath: "/file.txt"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Download error = %v, want context.Canceled", err)
	}
}

func TestClientWrapsProviderError(t *testing.T) {
	providerErr := errors.New("provider failed")
	client := New(WithProvider(&fakeProvider{err: providerErr}))
	_, err := client.Upload(context.Background(), UploadRequest{RemotePath: "/file.txt"})
	if !errors.Is(err, providerErr) {
		t.Fatalf("Upload error = %v, want provider error", err)
	}
}

func TestClientRunEnforcesOutputLimit(t *testing.T) {
	client := New(WithProvider(&fakeProvider{runResponse: CommandResponse{Stdout: []byte("hello"), Stderr: []byte("!")}}))
	_, err := client.Run(context.Background(), CommandRequest{Command: "printf", MaxOutputBytes: 5})
	if !errors.Is(err, ErrOutputLimitExceeded) {
		t.Fatalf("Run error = %v, want ErrOutputLimitExceeded", err)
	}
}

func TestClientDownloadEnforcesTransferLimit(t *testing.T) {
	client := New(WithProvider(&fakeProvider{downloadResponse: DownloadResponse{RemotePath: "/file.txt", Content: []byte("hello")}}))
	_, err := client.Download(context.Background(), DownloadRequest{RemotePath: "/file.txt", MaxBytes: 4})
	if !errors.Is(err, ErrTransferLimitExceeded) {
		t.Fatalf("Download error = %v, want ErrTransferLimitExceeded", err)
	}
}

func TestClientUploadEnforcesProviderReportedTransferLimit(t *testing.T) {
	client := New(WithProvider(&fakeProvider{uploadResponse: UploadResponse{RemotePath: "/file.txt", Size: 6}}))
	_, err := client.Upload(context.Background(), UploadRequest{RemotePath: "/file.txt", Content: []byte("hello"), MaxBytes: 5})
	if !errors.Is(err, ErrTransferLimitExceeded) {
		t.Fatalf("Upload error = %v, want ErrTransferLimitExceeded", err)
	}
}

func TestClientUploadClonesContent(t *testing.T) {
	provider := &fakeProvider{uploadResponse: UploadResponse{RemotePath: "/file.txt", Size: 5}}
	client := New(WithProvider(provider))
	content := []byte("hello")
	response, err := client.Upload(context.Background(), UploadRequest{RemotePath: "/file.txt", Content: content})
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	content[0] = 'H'
	if !reflect.DeepEqual(provider.uploadRequests[0].Content, []byte("hello")) {
		t.Fatalf("provider upload content was not cloned: %q", provider.uploadRequests[0].Content)
	}
	if response.Size != 5 {
		t.Fatalf("Upload response = %+v", response)
	}
}
