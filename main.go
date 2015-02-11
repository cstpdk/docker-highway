package main

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
	"flag"
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
	HostNumber  int
	HostValue   string
	HostPort    int
}

func (ee *EtcdHostEntry) fromContainer(container Container) *EtcdHostEntry {

	ee.ServiceName = container.Names[0]
	ee.Scheme = "http"

	if len(container.Ports) > 0 {
		ee.HostNumber = 1
		ee.HostValue = fmt.Sprintf("%s:%d", container.Ports[0].IP,
			container.Ports[0].PublicPort)
	}

	return ee
}

func (ee *EtcdHostEntry) saveOrUpdate(conn *etcd.Client) {
	_, err := conn.Set(fmt.Sprintf("/services/%s/hosts/%d", ee.ServiceName,
		ee.HostNumber),
		ee.HostValue, 0)

	handleError(err)

	_, err = conn.Set(fmt.Sprintf("/services/%s/scheme", ee.ServiceName),
		ee.Scheme, 0)

	handleError(err)
}

func handleError(err error) {
	if err != nil {
		println(err.Error())
	}
}

func get(dockerConn *httputil.ClientConn, path string) ([]byte, error){

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
		body, err := get(dockerConn,"/containers/json")
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
	endpoint := fmt.Sprintf("http://%s",flag.Arg(0))
	etcdConn := etcd.NewClient([]string{endpoint})

	listenForContainers(dockerConn, etcdConn)
}
