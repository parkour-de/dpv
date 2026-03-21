package graph

import (
	"context"
	"dpv/dpv/src/repository/t"

	"github.com/arangodb/go-driver/v2/arangodb"
)

type EntityManager[T Entity] struct {
	Collection  arangodb.Collection
	Constructor func() T
	Audit       *AuditLogger
	Type        string
}

type Entity interface {
	GetKey() string
	SetKey(key string)
}

func (im *EntityManager[T]) Create(item T, ctx context.Context) error {
	meta, err := im.Collection.CreateDocument(ctx, item)
	if err != nil {
		return t.Errorf("could not create item: %w", err)
	}
	item.SetKey(meta.Key)
	if im.Audit != nil {
		im.Audit.Log(ctx, ActionCreate, im.Type, meta.Key, item)
	}
	return nil
}

func (im *EntityManager[T]) Has(key string, ctx context.Context) (bool, error) {
	exists, err := im.Collection.DocumentExists(ctx, key)
	if err != nil {
		return false, t.Errorf("could not check for item with key %v: %w", key, err)
	}
	return exists, nil
}

func (im *EntityManager[T]) Read(key string, ctx context.Context) (T, error) {
	item := im.Constructor()
	meta, err := im.Collection.ReadDocument(ctx, key, item)
	if err != nil {
		return item, t.Errorf("could not read item with key %v: %w", key, err)
	}
	item.SetKey(meta.Key)
	return item, nil
}

func (im *EntityManager[T]) Update(item T, ctx context.Context) error {
	_, err := im.Collection.UpdateDocument(ctx, item.GetKey(), item)
	if err != nil {
		return t.Errorf("could not update item with key %v: %w", item.GetKey(), err)
	}
	if im.Audit != nil {
		im.Audit.Log(ctx, ActionUpdate, im.Type, item.GetKey(), item)
	}
	return nil
}

func (im *EntityManager[T]) Patch(key string, patchObject map[string]interface{}, ctx context.Context) error {
	query := "UPDATE @key WITH @patch IN @@collection OPTIONS { keepNull: false }"
	bindVars := map[string]interface{}{
		"key":         key,
		"patch":       patchObject,
		"@collection": im.Collection.Name(),
	}
	_, err := im.Collection.Database().Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
	if err != nil {
		return t.Errorf("could not patch item with key %v: %w", key, err)
	}
	if im.Audit != nil {
		im.Audit.Log(ctx, ActionUpdate, im.Type, key, patchObject)
	}
	return nil
}

func (im *EntityManager[T]) Delete(item T, ctx context.Context) error {
	_, err := im.Collection.DeleteDocument(ctx, item.GetKey())
	if err != nil {
		return t.Errorf("could not delete item with key %v: %w", item.GetKey(), err)
	}
	if im.Audit != nil {
		im.Audit.Log(ctx, ActionDelete, im.Type, item.GetKey(), nil)
	}
	return nil
}
