package main

import (
	"fmt"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/juliengk/go-utils/filedir"
)

type nfsVolume struct {
	Server  string
	Path    string
	Options string

	Mountpoint  string
	connections int
}

func (d *volumeDriver) Create(req volume.Request) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Create: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	_, ok := req.Options["server"]
	if !ok {
		res.Err = fmt.Sprintf("server option is mandatory")
		return res
	}
	if len(req.Options["server"]) == 0 {
		res.Err = fmt.Sprintf("server cannot be empty")
		return res
	}

	_, ok = req.Options["path"]
	if !ok {
		res.Err = fmt.Sprintf("path option is mandatory")
		return res
	}
	if len(req.Options["path"]) == 0 {
		res.Err = fmt.Sprintf("path cannot be empty")
		return res
	}

	server, err := HostLookup(req.Options["server"])
	if err != nil {
		res.Err = err.Error()
		return res
	}

	vol := nfsVolume{
		Server:     server,
		Path:       req.Options["path"],
		Mountpoint: d.getPath(req.Name),
	}

	_, ok = req.Options["opts"]
	if ok {
		vol.Options = req.Options["opts"]
	}

	if err := d.addVolume(req.Name, &vol); err != nil {
		res.Err = err.Error()
		return res
	}

	d.saveState()

	return res
}

func (d *volumeDriver) List(req volume.Request) volume.Response {
	var res volume.Response

	log.Info("VolumeDriver.List: volumes")

	d.Lock()
	defer d.Unlock()

	res.Volumes = d.listVolumes()

	return res
}

func (d *volumeDriver) Get(req volume.Request) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Get: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	v, err := d.getVolume(req.Name)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	res.Volume = &volume.Volume{
		Name:       req.Name,
		Mountpoint: v.Mountpoint,
	}

	return res
}

func (d *volumeDriver) Remove(req volume.Request) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Remove: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	if err := d.removeVolume(req.Name); err != nil {
		res.Err = err.Error()
		return res
	}

	d.saveState()

	return res
}

func (d *volumeDriver) Path(req volume.Request) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Path: volume %s", req.Name)

	d.RLock()
	defer d.RUnlock()

	_, err := d.getVolume(req.Name)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	res.Mountpoint = d.getPath(req.Name)

	return res
}

func (d *volumeDriver) Mount(req volume.MountRequest) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Mount: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	v, err := d.getVolume(req.Name)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	if v.connections == 0 {
		if err := filedir.CreateDirIfNotExist(v.Mountpoint, true, 0700); err != nil {
			res.Err = err.Error()
			return res
		}

		source := fmt.Sprintf(":%s", v.Path)
		target := v.Mountpoint
		opts := []string{
			"nolock",
			fmt.Sprintf("addr=%s", v.Server),
		}
		options := strings.Join(opts, ",")

		if err := syscall.Mount(source, target, "nfs", 0, options); err != nil {
			res.Err = err.Error()
			return res
		}
	}

	v.connections++

	res.Mountpoint = v.Mountpoint

	return res
}

func (d *volumeDriver) Unmount(req volume.UnmountRequest) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Unmount: volume %s", req.Name)

	d.Lock()
	defer d.Unlock()

	v, err := d.getVolume(req.Name)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	v.connections--

	if v.connections <= 0 {
		if err := syscall.Unmount(v.Mountpoint, 0); err != nil {
			res.Err = err.Error()
			return res
		}

		v.connections = 0
	}

	return res
}

func (d *volumeDriver) Capabilities(req volume.Request) volume.Response {
	var res volume.Response

	log.Infof("VolumeDriver.Capabilities: volume %s", req.Name)

	res.Capabilities = volume.Capability{Scope: "local"}

	return res
}
