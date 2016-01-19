package main

import (
	"fmt"
	"github.com/GeertJohan/go.ask"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ++++++++++++++ Algorithm at a glance ++++++++++++++++++
// build base:
//   For each node:
//     Determine tags
//     Append node to nodes with lastrowid
//     for each Tag
//       append to tagIndex with this node_id
//       append tag to Node

// build tree:
// for each node in Nodes:
//   for each tag in sorted_by_popularity(tags):
//     append to tree

// query tree:
//   select first child.
//     query it.
//     if true:
//       drill_down()
//     else:
//       select second child

// parses yaml config: attrs.yaml file
func Parse_config() []Element_yaml {
	asset, err := Asset("attrs.yaml")
	if err != nil {
		// Asset was not found.
		log.Println(err)
	}
	// file, err := ioutil.ReadFile(CONFIG)
	// if err != nil {
	// 	log.Println(err)
	// }
	data := make([]Element_yaml, 0)

	err2 := yaml.Unmarshal(asset, &data)
	if err2 != nil {
		log.Println(err2)
	}
	return data
}

// assign tags to each file by calling appropriate workers(see extensions.go)
func (s *State) Worker(path string, p AttrPath, f os.FileInfo) {
	s.LastRowid += 1
	nodes_tags := make(map[Tag]bool)

	s.Nodes[s.LastRowid] = Node{
		Node:  path,
		Tags:  nodes_tags,
		Rowid: s.LastRowid,
	}
	for _, v := range s.Conf {
		// build tags List
		result := Dispatch_Tag(v, p, f)
		if result.Tag != "" {
			// append tag to Tag Index
			if s.TagIndex[result.Tag] == nil {
				tag_result := make(map[int]bool)
				tag_result[s.LastRowid] = true
				s.TagIndex[result.Tag] = tag_result
			} else {
				s.TagIndex[result.Tag][s.LastRowid] = true
			}
			// append tag to nodes
			nodes_tags[result.Tag] = true
			// append tag ask to asks
			s.TagAsk[result.Tag] = result.TagAsk
		}
	}
}

func (s *State) FSWalker(path string, info os.FileInfo, err error) error {
	x := AttrPath{}
	x.Base = filepath.Base(path)
	x.Dir = filepath.Dir(path)
	x.Ext = filepath.Ext(path)
	x.isAbs = filepath.IsAbs(path)
	s.Worker(path, x, info)
	s.FilesInDir += 1
	return nil
}
func (s *State) Walk() {
	filepath.Walk(s.Path, s.FSWalker)
}

func (s *State) Sort_By_Popularity(rowid int) *ValSorter {
	// agggrh ...  this language.
	// stay away. I a sorting
	tags_map := make(map[string]int)

	for k, _ := range s.Nodes[rowid].Tags {
		tags_map[string(k)] = len(s.TagIndex[k])
	}
	vs := NewValSorter(tags_map)
	vs.Sort()
	return vs
}

func (s *State) build_node_branch(rowid int) {
	vs := s.Sort_By_Popularity(rowid)
	vs.Sort()
	cur_el := s.Tree

	for _, v := range vs.Keys {
		if cur_el[Tag(v)] == nil {
			new_branch := make(TreeIndex)
			cur_el[Tag(v)] = &new_branch
		}
		cur_el = *cur_el[Tag(v)]
	}
}
func (s *State) build_tree() {
	for k, _ := range s.Nodes {
		s.build_node_branch(k)
	}
}

// return And result from akinate
func (s *State) get_and(data map[Tag]bool) map[string]bool {
	nodes := make(map[int]int)
	result := make(map[string]bool)
	for k, _ := range data {
		for k2, _ := range s.TagIndex[k] {
			nodes[k2] += 1
		}
	}
	for k, v := range nodes {
		if v == len(data) {
			result[s.Nodes[k].Node] = true
		}
	}
	return result
}

// display debug info
func (s *State) show_tagSizes() {
	for k, v := range s.TagIndex {
		log.Println(k, len(v))
	}
}

// display debug info
func (s *State) walk_tree(t TreeIndex, space string) {
	sortableEl := make(map[string]int)
	for k := range t {
		sortableEl[string(k)] = len(s.TagIndex[k])
	}
	vs := NewValSorter(sortableEl)
	vs.Sort()
	for _, v := range vs.Keys {
		log.Println(space, v)
		s.walk_tree(*t[Tag(v)], space+" ")
	}
}

func (s *State) here_are_your_files(data map[Tag]bool) {
	fmt.Println("Here are your files!")
	for k, _ := range s.get_and(data) {

		fmt.Println(k)
	}
	s.copyTo(data)
	s.moveTo(data)
	os.Exit(1)
}

func (s *State) moveTo(data map[Tag]bool) {
	if MOVE_RESULT_TO == "" {
		return
	}
	fmt.Println("Now, we are moving your files  to ", MOVE_RESULT_TO)
	err := os.MkdirAll(MOVE_RESULT_TO, 0777)
	if err != nil {
		log.Println("Cannot create directory: ", MOVE_RESULT_TO)
		log.Println("Error is ", err)
	}
	for k, _ := range s.get_and(data) {
		err := os.Rename(k, HandlePathClashes(filepath.Join(MOVE_RESULT_TO, filepath.Base(k))))
		if err != nil {
			log.Println("Cannot move file: ", k, " to directory: ", MOVE_RESULT_TO, " Error is: ", err)
		}
	}
}

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func HandlePathClashes(fpath string) string {
	// increment filename by counter e.g. main.css => main(1).css
	counter := 1
	for {
		if _, err := os.Stat(fpath); os.IsNotExist(err) {
			return fpath
		}
		base, fname := filepath.Split(fpath)
		fname_ext := filepath.Ext(fname)
		fname_base := strings.TrimRight(fname, fname_ext)
		new_fname := strings.Join([]string{fname_base, "(", strconv.Itoa(counter), ")", fname_ext}, "")
		fpath = filepath.Join(base, new_fname)
	}
}

func (s *State) copyTo(data map[Tag]bool) {
	if COPY_RESULT_TO == "" {
		return
	}
	fmt.Println("Now, we are copying  your files  to ", COPY_RESULT_TO)
	err := os.MkdirAll(COPY_RESULT_TO, 0777)
	if err != nil {
		log.Println("Cannot create directory: ", COPY_RESULT_TO)
		log.Println("Error is ", err)
	}
	for k, _ := range s.get_and(data) {
		// os.IsExist(err error)
		k = HandlePathClashes(k)
		err := Copy(k, COPY_RESULT_TO)
		if err != nil {
			log.Println("Cannot copy file: ", k, " to directory: ", COPY_RESULT_TO, " Error is: ", err)
		}
	}
}

func (s *State) Akinate(t TreeIndex, data map[Tag]bool) {
	sortableEl := make(map[string]int)
	if len(t) == 0 {
		s.here_are_your_files(data)
	}
	for k := range t {
		sortableEl[string(k)] = len(s.TagIndex[k])
	}
	vs := NewValSorter(sortableEl)
	vs.Sort()
	for _, v := range vs.Keys {
		result, _ := ask.Ask(s.TagAsk[Tag(v)])
		if result {
			data[Tag(v)] = true
			s.Akinate(*t[Tag(v)], data)
			// exiting level should be prevented
			s.here_are_your_files(data)
		}
	}
}
