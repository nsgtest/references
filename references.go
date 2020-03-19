package main

import(
	"fmt"
	"io/ioutil"
	"encoding/json"
	"os"
	"path"
	"crypto/sha256"
	"strings"
	"github.com/nsgtest/packages/interfaces"
	"github.com/nsgtest/packages/structs"
)

var messages []string = []string{
	"No similar reference found!",
	"Reference name already assigned:",
	"Something went horribly wrong:",
	"Found same reference with differenct content(SHA256):",
	"Found reference with same content(SHA256):",
	"Found same reference with different URL:",
	"Something went horribly wrong.",
	"Found identical reference:"}

type Reference struct{
	Name, Url	string
	Checksum	[32]byte
}

type References []Reference

func main(){
	if len(os.Args) > 1{
		switch os.Args[1]{
		case "init":
			if len(os.Args) == 3{
				s := structs.Struct{os.Args[2], Reference{}, References{}.ota()}
				s.Write()
			} else {
				help()
			}
		case "add":
			if len(os.Args) == 4{
				s := structs.Struct{os.Args[2], Reference{path.Base(os.Args[3]), url(os.Args[3]), checksum(os.Args[3])}, read(os.Args[2]).ota()}
				s.Add([]int{1,2,3,4,5,6,7})
			} else {
				help()
			}
		case "update":
			if len(os.Args) == 4{
				s := structs.Struct{os.Args[2], Reference{path.Base(os.Args[3]), url(os.Args[3]), checksum(os.Args[3])}, read(os.Args[2]).ota()}
				s.Update([]int{3,5})
			} else {
				help()
			}
		case "remove":
			if len(os.Args) == 4{
				s := structs.Struct{os.Args[2], Reference{os.Args[3], "", [32]byte{}}, read(os.Args[2]).ota()}
				s.Remove([]int{1})
			} else {
				help()
			}
		case "list":
			if len(os.Args) == 3{
				s := structs.Struct{os.Args[2], Reference{}, read(os.Args[2]).ota()}
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

func (r Reference) Message(count int){
	fmt.Printf(messages[count])
}

func (r References) ota() interfaces.Interfaces{
	array := make(interfaces.Interfaces, len(r))
	for i, reference := range r{
		array[i] = reference
	}
	return array
}

func checksum(file string) [32]byte{
	data, err := ioutil.ReadFile(file)
	if err != nil{
		fmt.Printf("FAIL!\nCould not read from %v!\n\n", file)
		panic(err)
	}

	return sha256.Sum256(data)
}

func help(){
	fmt.Printf("Usage: refs COMMAND [ARGUMENT]\nCreate references and add, modify or remove a reference.\n\nCOMMANDS:\n\tnew NAME\t\tCreate new refernces named NAME.\n\tadd NAME FILE\t\tAdd reference FILE to references NAME.\n\tupdate NAME FILE\tUpdate reference FILE in reference NAME.\n\tremove NAME REFERENCE\tRemove REFERENCE from reference NAME.\n\tlist NAME\t\tList references NAME.\n\nARGUMENTS:\n\tNAME\t\tValid JSON file name.\n\tFILE\t\tAbsolute Path to file.\n\tREFERENCE\tName of a reference in references.\n")
}

func read(file string) References{
	enc, err := ioutil.ReadFile(file)
	if err != nil{
		fmt.Printf("FAIL!\nCould not read from %v!\n\n", file)
		panic(err)
	}

	r := References{}
	err = json.Unmarshal(enc, &r)
	if err != nil{
		fmt.Printf("FAIL!\n%v is not a JSON file!\n\n", file)
		panic(err)
	}

	return r
}

func url(file string) string{
	_, err := os.Stat(file)
	if os.IsNotExist(err){
		fmt.Printf("FAIL!\n%v does not exist!\n\n", file)
		panic (err)
	}

	if !path.IsAbs(file){
		fmt.Printf("FAIL!\n%v is not an absolute file path!\n", file)
		panic(nil)
	}

	content := path.Base(file)
	file = path.Dir(file)

	_, err = os.Stat(path.Join(file, ".git"))
	for os.IsNotExist(err){
		content = path.Join(path.Base(file), content)
		file = path.Dir(file)

		if file == "/" {
			fmt.Printf("FAIL\n%v is not a git repository!\n", file)
			panic(err)
		}

		_, err = os.Stat(path.Join(file, ".git"))
	}
	data, err := ioutil.ReadFile(path.Join(file, ".git/logs/refs/remotes/origin/HEAD"))
	if err != nil{
		fmt.Printf("FAIL!\nCould not read from %v!\n\n", file)
		panic(err)
	}

	split := strings.Split(string(data), " ")
	url := strings.TrimSuffix(split[len(split) - 1], "\n")

	repository := strings.TrimSuffix(path.Base(url), path.Ext(url))
	url = path.Dir(url)
	repository = path.Join(path.Base(url), repository)

	return path.Join("https://api.github.com/repos", repository, "contents", content + "?ref=master")
}
