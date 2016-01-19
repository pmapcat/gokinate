package main

import (
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type AttrPath struct {
	Base  string
	Dir   string
	Ext   string
	isAbs bool
}

func (a *AttrPath) Path() string {
	return filepath.Join(a.Dir, a.Base)
}

func (el *Element_yaml) ext_worker(p AttrPath, i os.FileInfo) WorkerReturn {
	data := WorkerReturn{}
	for _, v := range el.Items {
		if v == p.Ext {
			data.Tag = Tag(el.Yes)
			data.TagAsk = el.Ask
			return data
		}
	}
	return data
}

func (el *Element_yaml) path_substring_worker(p AttrPath, i os.FileInfo) WorkerReturn {
	data := WorkerReturn{}
	fpath := p.Path()
	for _, v := range el.Items {
		if strings.Contains(fpath, v) {
			data.Tag = Tag(el.Yes)
			data.TagAsk = el.Ask
			return data
		}
	}
	return data
}

func (el *Element_yaml) date_worker(p AttrPath, i os.FileInfo) WorkerReturn {
	data := WorkerReturn{}
	date := time.Now()
	delta := date.Sub(i.ModTime())
	deltaHours := delta.Hours()

	min, min_err := strconv.Atoi(el.Items[0])
	if min_err != nil {
		log.Println("Canot parse string as number")
		return data
	}
	max, max_err := strconv.Atoi(el.Items[1])
	if max_err != nil {
		log.Println("Canot parse string as number")
		return data
	}
	if (float64(min) < deltaHours) && (deltaHours < float64(max)) {
		data.Tag = Tag(el.Yes)
		data.TagAsk = el.Ask
	}
	return data
}

func (el *Element_yaml) size_range_worker(p AttrPath, i os.FileInfo) WorkerReturn {
	data := WorkerReturn{}
	size := float64(i.Size())

	min, min_err := strconv.Atoi(el.Items[0])
	if min_err != nil {
		log.Println("Canot parse string as number")
		return data
	}
	max, max_err := strconv.Atoi(el.Items[1])
	if max_err != nil {
		log.Println("Canot parse string as number")
		return data
	}
	minSize := float64(min) * math.Pow(10, 6)
	maxSize := float64(max) * math.Pow(10, 6)

	if (minSize < size) && (size < maxSize) {
		data.Tag = Tag(el.Yes)
		data.TagAsk = el.Ask
	}
	return data
}

// dispatch workers, so you can
func Dispatch_Tag(el Element_yaml, p AttrPath, f os.FileInfo) WorkerReturn {
	result := WorkerReturn{}
	if el.Worker == "ext_worker" {
		result = el.ext_worker(p, f)
	}
	if el.Worker == "date_worker" {
		result = el.date_worker(p, f)
	}
	if el.Worker == "size_range" {
		result = el.size_range_worker(p, f)
	}
	if el.Worker == "path_substring_worker" {
		result = el.path_substring_worker(p, f)
	}
	return result
}
