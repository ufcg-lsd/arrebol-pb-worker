package utils

//This file implements some functions that are usually called in sequence
//to achieve some common results, some of them are listed below:
//Create a container and let it ready: CheckImage; Pull; CreateContainer; StartContainer.
//Copy a file from the host to the container: Copy.
//To write some array of content to a file inside the container: Write.
//To run a valid command inside the container: Exec
//To kill/remove the container: StopContainer; RemoveContainer.
//Note that the sequence above is usually ran to use the container for the most common purposes.

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type ContainerConfig struct {
	Name   string
	Image  string
	Mounts []mount.Mount
}

//Creates a new docker client
//Params:
//host - the host address in which the client
//will be created. (e.g 127.0.0.1:5555)
//It returns:
//1. nil if the host pattern is invalid
//2. a docker client otherwise
func NewDockerClient(host string) *client.Client {
	if err := os.Setenv("DOCKER_HOST", host); err != nil {
		log.Print(err)
		return nil
	}
	log.Println("Starting docker client in host: " + host)
	cli, err := client.NewEnvClient()

	if err != nil {
		log.Println(err)
	}

	return cli
}

//Creates a container
//Params:
//cli - the docker client whose host will get the new container
//config - the container configuration. That's the way to set
//the container name, image and possible mounts.
//It returns:
//1. an empty string and an error if it faces some problem on container creation
//(e.g a already used container name)
//2. the container id and nil otherwise.
func CreateContainer(cli *client.Client, config ContainerConfig) (string, error) {
	log.Printf("Creating Container [%s]", config.Name)
	ctx := context.Background()
	hostConfig := container.HostConfig{
		Mounts: config.Mounts,
	}

	dconfig := container.Config{
		Image: config.Image,
		Tty:   true,
	}

	b, err := cli.ContainerCreate(ctx, &dconfig, &hostConfig, nil, config.Name)

	if err != nil {
		log.Println(err)
	}

	return b.ID, err
}

//Starts an existent container
//Params:
//cli - the docker client
//id - the container id
//It returns:
//1. an error if the passed id doesn't exists
//2. nil otherwise.
func StartContainer(cli *client.Client, id string) error {
	log.Printf("Starting Container [%s]", id)
	return cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
}

//Stops a container
//Params:
//cli - the docker client
//id - the container id
//It returns:
//1. an error if the passed id doesn't exists
//2. nil otherwise.
func StopContainer(cli *client.Client, id string) error {
	log.Printf("Stopping Container [%s]", id)
	var timeout = 5 * time.Second
	return cli.ContainerStop(context.Background(), id, &timeout)
}

//Removes a container
//Params:
//cli - the docker client
//id - the container id
//It returns:
//1. an error if the passed id doesn't exists
//2. nil otherwise.
func RemoveContainer(cli *client.Client, id string) error {
	log.Printf("Removing Container [%s]", id)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{})
}

//Iterates over the content and write each one to the destination file inside the container
//Params:
//cli - the docker client
//id - the container id
//content - the array that stores the content.
//Each position of the array will become a line in the dest file
//dest - the destination file path, inside the container.
//It returns:
//1. an error if the passed id doesn't exists or if the destination file is a invalid one
//2. nil otherwise.
func Write(cli *client.Client, id string, content []string, dest string) error {
	for _, c := range content {
		c = strings.ReplaceAll(c, "'", "'\"'\"'")
		cmd := fmt.Sprintf(`echo -E '%s' >> %s`, c, dest)
		log.Printf("Writing [%s] on [%s] from Container [%s]", c, dest, id)
		err := Exec(cli, id, cmd)

		if err != nil {
			return err
		}
	}
	return nil
}

//It copies the src file, which lives in the worker host,
//to the dest file inside the container.
//Params:
//cli - the docker client
//id - the container id
//src - the source file path (in the worker host)
//dest - the destination file, inside the container
//It returns:
//1. an error if the passed id doesn't exists or if the destination file is a invalid one
//2. nil otherwise.
func Copy(cli *client.Client, id, src, dest string) error {
	log.Printf("Copy [%s] to [%s] from Container [%s]", src, dest, id)
	dat, _ := ioutil.ReadFile(src)
	content := string(dat)
	content = strings.ReplaceAll(content, "'", "'\"'\"'")
	cmd := fmt.Sprintf("echo -E '%s' >| %s", content, dest)
	return Exec(cli, id, cmd)
}

//Executes a bash command inside the container
//Params:
//cli - the docker client
//id - the container id
//cmd - the bash command (e.g "echo 'arrebol'")
//It returns:
//1. an error if the command couldn't be executed inside the container
//(e.g call a binary that doesn't exists), or if the id doesn't exists
//2. nil otherwise.
func Exec(cli *client.Client, id, cmd string) error {
	log.Printf("Executing command [%s] on container [%s]", cmd, id)
	config := types.ExecConfig{
		Cmd: []string{"/bin/bash", "-c", cmd},
	}
	rid, _ := cli.ContainerExecCreate(context.Background(), id, config)
	return cli.ContainerExecStart(context.Background(), rid.ID, types.ExecStartCheck{})
}

//Executes cat in a file inside the container and returns its output
//Params:
//cli - the docker client
//id - the container id
//path - the file path inside the container
//It returns:
//1. nil and an error if the id doesn't exists,
//or if the file path is invalid.
//2. The file content as byte array and nil otherwise.
func Read(cli *client.Client, id, path string) ([]byte, error) {
	log.Printf("Getting content of file [%s]", path)
	config := types.ExecConfig{
		Tty:          true,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          []string{"/bin/bash", "-c", "cat " + path},
	}
	rid, err := cli.ContainerExecCreate(context.Background(), id, config)
	if err != nil {
		log.Println("error on creating container exec")
		return nil, err
	}
	log.Println(rid.ID)
	hijack, err := cli.ContainerExecAttach(context.Background(), rid.ID, config)

	if err != nil {
		return nil, err
	}
	output, err := read(hijack.Conn)

	if err != nil {
		return nil, err
	}
	return output, nil
}

//It reads the content that comes from the connection received
//as parameter
//Params:
//conn - the connection from which the content is read
//It returns:
//1. nil and an error, if there is no content in the connection socket
//2. a byte array with the content that came from the connection and nil otherwise
func read(conn net.Conn) ([]byte, error) {
	result := make([]byte, 0)
	b := make([]byte, 10)
	for {
		n, err := conn.Read(b)

		if err != nil {
			log.Println(err)
			return nil, err
		}

		result = append(result, b...)
		if n < len(b) {
			break
		}
		b = make([]byte, 2*len(b))
	}
	return result, nil
}

//Downloads a docker image
//Params:
//cli - the docker client
//image - the docker image (e.g library/ubuntu:16.04)
//It returns:
//1. nil and an error if the image couldn't be downloaded
//2. A reader and nil otherwise.
func Pull(cli *client.Client, image string) (io.ReadCloser, error) {
	reader, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	return reader, err
}

//Checks if the image is valid.
//In the image library/ubuntu:16.04, for example, it checks if
//library/ubuntu really exists, and if 16.04 is a valid tag.
//Params:
//cli - docker client
//image - the docker image
//It returns:
//1. false and an error if the image is invalid
//2. true and nil otherwise
func CheckImage(cli *client.	Client, image string) (exist bool, err error) {
	exist = false
	_, _, err = cli.ImageInspectWithRaw(context.Background(), image)
	if err == nil {
		exist = true
	}
	return
}
