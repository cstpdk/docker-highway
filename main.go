package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type Container struct {
	Command string
	Created int64
	Id      string
	Image   string
	Names   []string
	Ports   []Port
	Status  string
}

type Port struct {
	IP          string
	PrivatePort int
	PublicPort  int
	Type        string
}

type EtcdHostEntry struct {
	Scheme      string
	ServiceName string
	HostName    string //Name of this instance, like /services/ServiceName/HostName
	HostValue   string
	HostPort    int
}

func (ee *EtcdHostEntry) fromContainer(container Container) *EtcdHostEntry {

	// We split on "_" to allow for several containers with same name to run
	name_tokens := strings.Split(container.Names[0], "_")

	// Everything before "_" is the name
	ee.ServiceName = name_tokens[0]

	if len(name_tokens) > 1 {
		// If there is more to the name, use that
		ee.HostName = name_tokens[1]
	} else {
		ee.HostName = "1"
	}

	// Http is the default, can be overriden without being updated
	ee.Scheme = "http"

	if len(container.Ports) > 0 {
		ee.HostValue = fmt.Sprintf("%s:%d", container.Ports[0].IP,
			container.Ports[0].PublicPort)
	}

	return ee
}

func (ee *EtcdHostEntry) saveOrUpdate(conn *etcd.Client) {
	_, err := conn.Set(fmt.Sprintf("/services/%s/hosts/%s", ee.ServiceName,
		ee.HostName),
		ee.HostValue, 30)

	handleError(err)

	// Only suceeds if not already existant
	conn.Create(fmt.Sprintf("/services/%s/scheme", ee.ServiceName), ee.Scheme, 0)

	// TODO handle error if not "already exists"
}

func handleError(err error) {
	if err != nil {
		println(err.Error())
	}
}

func get(dockerConn *httputil.ClientConn, path string) ([]byte, error) {

	req, err := http.NewRequest("GET", path, nil)
	handleError(err)

	resp, err := dockerConn.Do(req)
	handleError(err)

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	return body, err
}

func listenForContainers(dockerConn *httputil.ClientConn, etcdConn *etcd.Client) {
	for {
		body, err := get(dockerConn, "/containers/json")
		handleError(err)

		containers := []Container{}
		json.Unmarshal(body, &containers)

		for _, container := range containers {
			ee := EtcdHostEntry{}
			ee.fromContainer(container).saveOrUpdate(etcdConn)
		}

		time.Sleep(5 * time.Second)
	}
}

func main() {

	dial, err := net.Dial("unix", "/var/run/docker.sock")
	handleError(err)
	defer dial.Close()

	dockerConn := httputil.NewClientConn(dial, nil)
	defer dockerConn.Close()

	flag.Parse()
	endpoint := fmt.Sprintf("http://%s", flag.Arg(0))
	etcdConn := etcd.NewClient([]string{endpoint})

	listenForContainers(dockerConn, etcdConn)
}
