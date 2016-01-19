package main

import (
	"flag"
	"log"
)

var MOVE_RESULT_TO = ""
var COPY_RESULT_TO = ""

type Tag string

type Node struct {
	Node  string
	Tags  map[Tag]bool
	Rowid int
}

type Index map[Tag]map[int]bool
type TreeIndex map[Tag]*TreeIndex

type State struct {
	FilesInDir int
	Nodes      map[int]Node
	Path       string
	TagIndex   Index
	LastRowid  int
	Tree       TreeIndex
	TagAsk     map[Tag]string
	Conf       []Element_yaml
}

func (s *State) init(path string) {
	s.Nodes = make(map[int]Node)
	s.Path = path
	s.TagIndex = make(map[Tag]map[int]bool, 0)
	s.Tree = make(TreeIndex, 0)
	s.TagAsk = make(map[Tag]string, 0)
	s.Conf = Parse_config()
}

type Element_yaml struct {
	Worker string   `yaml:"Worker"`
	Ask    string   `yaml:"Ask"`
	Yes    string   `yaml:"Yes"`
	Items  []string `yaml:"Items"`
}
type WorkerReturn struct {
	Tag    Tag
	TagAsk string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	s := State{}
	path := flag.String("p", "./", "Enter a dir, that you want to akinate")
	temp_mv := flag.String("mv", "", "Place a folder name where you want to move result")
	temp_cp := flag.String("cp", "", "Place a folder name where you want to copy result")
	is_debug := flag.Bool("debug", false, "Print debug info")
	flag.Parse()

	MOVE_RESULT_TO = *temp_mv
	COPY_RESULT_TO = *temp_cp
	s.init(*path)
	s.Walk()
	s.build_tree()

	if *is_debug {
		s.walk_tree(s.Tree, "-")
	}
	result_set := make(map[Tag]bool)
	s.Akinate(s.Tree, result_set)
}
