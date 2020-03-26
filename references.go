package main

import(
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"crypto/sha256"
	"strings"
	"github.com/nsgtest/packages/structs"
)

type Reference struct{
	Name, Url	string
	Checksum	[32]byte
}

func main(){
	if len(os.Args) > 1{
		switch os.Args[1]{
		case "init":
			if len(os.Args) == 3{
				s := structs.Struct{os.Args[2], Reference{}}
				s.Write()
			} else {
				help()
			}
		case "add":
			if len(os.Args) == 4{
				s := structs.Struct{os.Args[2], Reference{path.Base(os.Args[3]), url(os.Args[3]), checksum(os.Args[3])}}
				s.Add([]int{0})
			} else {
				help()
			}
		case "update":
			if len(os.Args) == 4{
				s := structs.Struct{os.Args[2], Reference{path.Base(os.Args[3]), url(os.Args[3]), checksum(os.Args[3])}}
				s.Update([]int{3,5,6,7})
			} else {
				help()
			}
		case "remove":
			if len(os.Args) == 4{
				s := structs.Struct{os.Args[2], Reference{os.Args[3], "", [32]byte{}}}
				s.Remove([]int{1})
			} else {
				help()
			}
		case "list":
			if len(os.Args) == 3{
				s := structs.Struct{os.Args[2], Reference{}}
				s.List()
			} else {
				help()
			}
		default:
			help()
		}
	} else {
		help()
	}
}

func (r Reference) Interface(){
	fmt.Println("I am the walrus!")
}

func checksum(file string) [32]byte{
	data, err := ioutil.ReadFile(file)
	if err != nil{
		fmt.Println("FAIL!")
		fmt.Printf("Could not read from %v!\n", file)
		panic(err)
	}

	return sha256.Sum256(data)
}

func help(){
	fmt.Printf("Usage: refs COMMAND [ARGUMENT]\nCreate references and add, modify or remove a reference.\n\nCOMMANDS:\n\tnew NAME\t\tCreate new refernces named NAME.\n\tadd NAME FILE\t\tAdd reference FILE to references NAME.\n\tupdate NAME FILE\tUpdate reference FILE in reference NAME.\n\tremove NAME REFERENCE\tRemove REFERENCE from reference NAME.\n\tlist NAME\t\tList references NAME.\n\nARGUMENTS:\n\tNAME\t\tValid JSON file name.\n\tFILE\t\tAbsolute Path to file.\n\tREFERENCE\tName of a reference in references.\n")
}

func url(file string) string{
	_, err := os.Stat(file)
	if os.IsNotExist(err){
		fmt.Println("FAIL!")
		fmt.Printf("%v does not exist!\n", file)
		panic (err)
	}

	if !path.IsAbs(file){
		fmt.Println("FAIL!")
		fmt.Printf("%v is not an absolute file path!\n", file)
		panic(nil)
	}

	content := path.Base(file)
	file = path.Dir(file)

	_, err = os.Stat(path.Join(file, ".git"))
	for os.IsNotExist(err){
		content = path.Join(path.Base(file), content)
		file = path.Dir(file)

		if file == "/" {
			fmt.Println("FAIL!")
			fmt.Printf("%v is not a git repository!\n", file)
			panic(err)
		}

		_, err = os.Stat(path.Join(file, ".git"))
	}
	data, err := ioutil.ReadFile(path.Join(file, ".git/logs/refs/remotes/origin/HEAD"))
	if err != nil{
		fmt.Println("FAIL!")
		fmt.Printf("Could not read from %v!\n", file)
		panic(err)
	}

	split := strings.Split(string(data), " ")
	url := strings.TrimSuffix(split[len(split) - 1], "\n")

	repository := strings.TrimSuffix(path.Base(url), path.Ext(url))
	url = path.Dir(url)
	repository = path.Join(path.Base(url), repository)

	return path.Join("https://api.github.com/repos", repository, "contents", content + "?ref=master")
}
