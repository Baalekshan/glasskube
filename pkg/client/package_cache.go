package client

import (
	"context"
	"errors"

	"github.com/glasskube/glasskube/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
)

type packageCacheClient struct {
	PackageV1Alpha1Client
	packageStore     cache.Store
	packageInfoStore cache.Store
}

func (c *packageCacheClient) Packages() PackageInterface {
	if c.packageStore == nil {
		return c.PackageV1Alpha1Client.Packages()
	}
	return &chachedPackageClient{PackageInterface: c.PackageV1Alpha1Client.Packages(), store: c.packageStore}
}

func (c *packageCacheClient) PackageInfos() PackageInfoInterface {
	if c.packageInfoStore == nil {
		return c.PackageV1Alpha1Client.PackageInfos()
	}
	return &cachedPackageInfoClient{
		PackageInfoInterface: c.PackageV1Alpha1Client.PackageInfos(),
		store:                c.packageInfoStore,
	}
}

type chachedPackageClient struct {
	PackageInterface
	store cache.Store
}

func (c *chachedPackageClient) Get(ctx context.Context, pkgName string, result *v1alpha1.Package) error {
	if obj, ok, err := c.store.GetByKey(pkgName); err != nil {
		return apierrors.NewInternalError(err)
	} else if !ok {
		return c.PackageInterface.Get(ctx, pkgName, result)
	} else if pkg, ok := obj.(*v1alpha1.Package); !ok {
		return apierrors.NewInternalError(errors.New("not a package"))
	} else {
		*result = *pkg
		return nil
	}
}

func (c *chachedPackageClient) GetAll(ctx context.Context, result *v1alpha1.PackageList) error {
	objs := c.store.List()
	items := make([]v1alpha1.Package, len(objs))
	for i, obj := range objs {
		if pkg, ok := obj.(*v1alpha1.Package); !ok {
			return apierrors.NewInternalError(errors.New("not a package"))
		} else {
			items[i] = *pkg
		}
	}
	*result = v1alpha1.PackageList{Items: items}
	return nil
}
