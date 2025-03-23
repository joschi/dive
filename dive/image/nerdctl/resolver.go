package nerdctl

import (
	"fmt"
	"io"

	"github.com/joschi/dive/dive/image"
	"github.com/joschi/dive/dive/image/docker"
)

type resolver struct{}

func NewResolverFromEngine() *resolver {
	return &resolver{}
}

// Name returns the name of the resolver to display to the user.
func (r *resolver) Name() string {
	return "nerdctl"
}

func (r *resolver) Build(args []string) (*image.Image, error) {
	id, err := buildImageFromCli(args)

	if err != nil {
		return nil, err
	}

	return r.Fetch(id)
}

func (r *resolver) Fetch(id string) (*image.Image, error) {
	img, err := r.resolveFromDockerArchive(id)

	if err == nil {
		return img, err
	}

	return nil, fmt.Errorf("unable to resolve image '%s': %+v", id, err)
}

func (r *resolver) Extract(id string, l string, p string) error {
	reader, err := streamNerdctlCmd("image", "save", id)

	if err != nil {
		return err
	}

	err = docker.ExtractFromImage(io.NopCloser(reader), l, p)

	if err != nil {
		fmt.Println("Handler not available locally. Trying to pull '" + id + "'...")
		err = runNerdctlCmd("pull", id)

		if err != nil {
			return err
		}

		reader, err = streamNerdctlCmd("image", "save", id)

		if err != nil {
			return err
		}

		err = docker.ExtractFromImage(io.NopCloser(reader), l, p)

		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("unable to extract from image '%s': %+v", id, err)
}

func (r *resolver) resolveFromDockerArchive(id string) (*image.Image, error) {
	reader, err := streamNerdctlCmd("image", "save", id)

	if err != nil {
		return nil, err
	}

	img, err := docker.NewImageArchive(io.NopCloser(reader))

	if err != nil {
		fmt.Println("Handler not available locally. Trying to pull '" + id + "'...")
		err = runNerdctlCmd("pull", id)

		if err != nil {
			return nil, err
		}

		reader, err = streamNerdctlCmd("image", "save", id)

		if err != nil {
			return nil, err
		}

		img, err = docker.NewImageArchive(io.NopCloser(reader))

		if err != nil {
			return nil, err
		}
	}

	return img.ToImage()
}
