# What is it, actually?
## First one
[gif about the docs]

## Second one
[gif about the nginx]
## Third one
[gif about the music]

# Performance
Examples in the gif(in downloads folder) above run in real time(no persistent quering) on,
approximately,  *50 000 files on traditional HDD*.
Basically, speed of collecting data and building a query tree is on par with ```Walk``` function in GoLang.

Internally *gokinate* uses O(N) *most prominent faceting* tree.
Which is, basically, a tree alignment based on most probable order of  *leafs/tags* in a node.

If made right, this tree can allow constant time *faceting/drilling down* in faceted drilling applications.
Index(tree) building is O(N) where N is a number of nodes.
We discard here sorting as it is a very small amount(5-10 tags to sort per node).

# Installation guide

```bash
go get github.com/mik/blablabblab;
cd $GOPATH/src/[my-repo];
go build;
go install;
gokinate -p ~/Downloads/;
```
# Extension guide

## Altering config file

You can find config file here
```bash
cd $GOPATH/src/mik-src/;
```
Go to the program directory and change **attrs.yaml** file

### Compilling attrs.yaml

You need go-bindata to be installed.

```bash
cd $GOPATH/src/mik-src/;
go-bindata "./attrs.yaml";
go build;
go install;
```

There are four  workers available

### File extension worker

```yaml
-
  Worker: ext_worker
  Yes: image_file
  Ask: Is this an image?
  Items:
    - .tif
    - .tiff
    - .gif
    - .jpeg
    - .jpg
    - .nef
    - .png
    - .psd
    - .bmp
```
    
It works like that. You associate file extensions with Ask and answer function.
Ask & answer are arbitrary fields. They can even, clash with each other.

### Path substring worker

```yaml
-
  Worker: path_substring_worker
  Yes: downloaded
  Ask: Did you download it from the internet?
  Items: [Downloads,Download]
```

### Size range worker

```yaml
-
  Worker: size_range_worker
  Yes: small_file
  Ask: Is this file small(less than 100 megabytes)?
  Items:  [0,99]
```
### Date range worker


```yaml
-
  Worker: date_worker
  Yes: yesterday_file
  Ask: Did you get this file yesterday?
  Items:  [1,23]
```

## Writing new extensions

Please, keep in mind, that complex extensions will substantially
slow down walking on directory.

For example, I've removed id3 functionality
from this program because it took very long time to build an mp3 tree.

New extension must be placed in *extensions.go* file.
In format:

```golang
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
```

And registered in dispatch tag function:

```golang
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
```
