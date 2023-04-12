package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"
)

const (
	GO_VERSION         = "1.20"
	SYFT_VERSION       = "v0.76.0"
	GORELEASER_VERSION = "v1.16.2"
	APP_NAME           = "dagger-go-example-app"
	BUILD_PATH         = "dist"
)

var (
	err      error
	res      string
	is_local bool
)

func main() {

	// Set a global flag when running locally
	flag.BoolVar(&is_local, "local", false, "whether to run locally [global]")
	flag.Parse()

	// Parse the first argument as the task
	task := flag.Arg(0)

	// Check if a task argument is provided
	if len(task) == 0 {
		log.Fatalln("Missing argument. Expected either 'pull-request' or 'release'.")
	}

	// Check if the task argument is valid
	if task != "pull-request" && task != "release" {
		log.Fatalln("Invalid argument. Expected either 'pull-request' or 'release'.")
	}

	// Dagger client context
	ctx := context.Background()

	// Create a Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}

	// Always close the client when done or on error.
	defer func() {
		log.Printf("Closing Dagger client...")
		client.Close()
	}()

	log.Println("Connected to Dagger")

	// Run the corresponding task command.
	switch task {
	case "pull-request":
		res, err = pullrequest(ctx, client)
	case "release":
		res, err = release(ctx, client)
	}

	// Handle any errors that occurred during the task execution.
	if err != nil {
		// log.Fatalf("Error %s: %+v\n", task, err)
		panic(fmt.Sprintf("Error %s: %+v\n", task, err))

	}

	log.Println(res)
}

// Pull Request Task: Runs tests and builds the binary
//
// `example: go run ci/dagger.go [-local,-help] pull-request `
func pullrequest(ctx context.Context, client *dagger.Client) (string, error) {

	// Get the source code from host directory and exclude files
	directory := client.Host().Directory(".")

	// Create a go container with the source code mounted
	golang := client.Container().
		From(fmt.Sprintf("golang:%s-alpine", GO_VERSION)).
		WithMountedDirectory("/src", directory).WithWorkdir("/src").
		WithMountedCache("/go/pkg/mod", client.CacheVolume("gomod")).
		WithEnvVariable("CGO_ENABLED", "0")

	// Run go tests
	_, err := golang.WithExec([]string{"go", "test", "./..."}).
		Stderr(ctx)

	if err != nil {
		return "", err
	}

	log.Println("Tests passed successfully!")

	// Run gorelaser release to check if the binary compiles in all platforms
	goreleaser := goreleaserContainer(ctx, client, directory).WithExec([]string{"release", "--snapshot", "--clean"})

	// Return any errors from the goreleaser build
	_, err = goreleaser.Stderr(ctx)

	if err != nil {
		return "", err
	}

	// Export builds to the host when running locally
	if is_local {

		// Retrieve the dist directory from the container
		dist := goreleaser.Directory(BUILD_PATH)

		// Export the dist directory when running locally
		_, err = dist.Export(ctx, BUILD_PATH)

		if err != nil {
			return "", err
		}

		log.Printf("Exported %v to local successfully!", BUILD_PATH)

	}

	return "Pull-Request tasks completed successfully!", nil
}

// Release Task: Runs GoReleaser to creates a Github release
//
// `example: go run ci/dagger.go release`
func release(ctx context.Context, client *dagger.Client) (string, error) {

	// Get the source code from host directory
	directory := client.Host().Directory(".")

	// Run gorelaser release to check if the binary compiles in all platforms
	goreleaser := goreleaserContainer(ctx, client, directory).WithExec([]string{"--clean"})

	// Return any errors from the goreleaser build
	_, err = goreleaser.Stderr(ctx)

	if err != nil {
		return "", err
	}

	return "Release tasks completed successfully!", nil
}

// goreleaserContainer returns a goreleaser container with the syft binary mounted and GITHUB_TOKEN secret set
//
// `example: goreleaserContainer(ctx, client, directory).WithExec([]string{"build"})`
func goreleaserContainer(ctx context.Context, client *dagger.Client, directory *dagger.Directory) *dagger.Container {
	// Set the Github token from the host environment as a secret
	token := client.SetSecret("github_token", os.Getenv("GITHUB_TOKEN"))

	// Export the syft binary from the syft container as a file
	syft := client.Container().From(fmt.Sprintf("anchore/syft:%s", SYFT_VERSION)).
		WithMountedCache("/go/pkg/mod", client.CacheVolume("gomod")).
		File("/syft")

	// Run go build to check if the binary compiles
	return client.Container().From(fmt.Sprintf("goreleaser/goreleaser:%s", GORELEASER_VERSION)).
		WithMountedCache("/go/pkg/mod", client.CacheVolume("gomod")).
		WithFile("/bin/syft", syft).
		WithMountedDirectory("/src", directory).WithWorkdir("/src").
		WithEnvVariable("TINI_SUBREAPER", "true").
		WithSecretVariable("GITHUB_TOKEN", token)

}
