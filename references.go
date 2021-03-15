package main

import(
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"crypto/sha256"
	"strings"
	"net/http"
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
			if len(os.Args) > 2{
				s := structs.Struct{os.Args[2], Reference{}}
				s.Write()
			} else {
				help()
			}
		case "add":
			if len(os.Args) > 3{
				s := structs.Struct{os.Args[2], Reference{filepath.Base(os.Args[3]), url(os.Args[3], upstream(os.Args)), checksum(os.Args[3])}}
				s.Add([]int{0})
			} else {
				help()
			}
		case "update":
			if len(os.Args) > 3{
				s := structs.Struct{os.Args[2], Reference{filepath.Base(os.Args[3]), url(os.Args[3], upstream(os.Args)), checksum(os.Args[3])}}
				s.Update([]int{3,5,6,7})
			} else {
				help()
			}
		case "remove":
			if len(os.Args) > 3{
				s := structs.Struct{os.Args[2], Reference{os.Args[3], "", [32]byte{}}}
				s.Remove([]int{1})
			} else {
				help()
			}
		case "list":
			if len(os.Args) > 2{
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
	fmt.Printf("Usage: refs COMMAND [OPTIONS]\nCreate references and add, modify or remove a reference.\n\nCOMMANDS:\n\tnew NAME\t\t\t\tCreate new refernces named NAME.\n\tadd NAME FILE [-u ACCOUNT]\t\tAdd reference FILE to references NAME.\n\tupdate NAME FILE [-u ACCOUNT]\t\tUpdate reference FILE in reference NAME.\n\tremove NAME REFERENCE\t\t\tRemove REFERENCE from reference NAME.\n\tlist NAME\t\t\t\tList references NAME.\n\nOPTIONS:\n\t-u, --upstream ACCOUNT\tUse Upstream ACCOUNT instead of Fork.\n\nARGUMENTS:\n\tNAME\t\tValid JSON file name.\n\tFILE\t\tAbsolute Path to file.\n\tREFERENCE\tName of a reference in references.\n\tACCOUNT\t\tExisting GitHub/GitLab Account.\n")
}

func upstream(args []string) *string{
	if len(os.Args) < 6{
		return nil;
	}

	for i, arg := range args[:len(args) - 1]{
		if arg == "-u" || arg == "--upstream" {
			resp, err := http.Head(url(os.Args[3], &args[i+1]))
			if err == nil && resp.StatusCode != http.StatusOK {
				fmt.Println("FAIL!")
				fmt.Printf("%v is not existing!\n", url(os.Args[3], &args[i+1]))
				panic(nil)
			}

			return &args[i+1]
		}
	}

	return nil
}

func url(file string, account *string) string{
	_, err := os.Stat(file)
	if os.IsNotExist(err){
		fmt.Println("FAIL!")
		fmt.Printf("%v does not exist!\n", file)
		panic(err)
	}

	if !filepath.IsAbs(file){
		fmt.Println("FAIL!")
		fmt.Printf("%v is not an absolute file path!\n", file)
		panic(nil)
	}

	filename := filepath.Base(file)
	file = filepath.Dir(file)

	_, err = os.Stat(filepath.Join(file, ".git"))
	for os.IsNotExist(err){
		filename = strings.Join([]string{filepath.Base(file), filename}, "/")
		file = filepath.Dir(file)

		if filepath.Dir(file) == filepath.Dir(filepath.Dir(file)) {
			fmt.Println("FAIL!")
			fmt.Printf("%v is not a git repository!\n", file)
			panic(err)
		}

		_, err = os.Stat(filepath.Join(file, ".git"))
	}

	data, err := ioutil.ReadFile(filepath.Join(file, ".git", "logs", "refs", "remotes", "origin", "HEAD"))
	if err != nil{
		fmt.Println("FAIL!")
		fmt.Printf("Could not read from %v!\n", file)
		panic(err)
	}

	datasplit := strings.Split(string(data[:len(data) - 1]), " ")

	urlsplit := strings.Split(datasplit[len(datasplit) - 1], "/")
	urlsplit[len(urlsplit) - 1] = strings.TrimSuffix(urlsplit[len(urlsplit) - 1], filepath.Ext(urlsplit[len(urlsplit) - 1]))

	if account != nil {
		urlsplit[3] = *account
	}

	if strings.Contains(urlsplit[2], "github"){
		return strings.Join([]string{urlsplit[0], "", "raw.githubusercontent.com", strings.Join(urlsplit[3:], "/"), "master", filename}, "/")
	}

	if strings.Contains(urlsplit[2], "gitlab"){
		return strings.Join([]string{urlsplit[0], "", urlsplit[2], strings.Join(urlsplit[3:], "/"), "-", "raw", "master", strings.Join([]string{filename, "?inline=false"}, "")}, "/")
	}

	fmt.Println("FAIL!")
	fmt.Printf("%v is neither a github nor a gitlab repository!\n", strings.Join([]string{urlsplit[0], strings.Join(urlsplit[2:], "/")}, "//"))
	panic(nil)
}
