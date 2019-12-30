package blobstorage

import "context"

type BlobStorage interface {
	Save(ctx context.Context, fileName string, content []byte) error
	Load(ctx context.Context, fileName string) (content []byte, err error)
}
