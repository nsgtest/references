package main

import(
	"encoding/json"
	"io/ioutil"
	"crypto/sha256"
	"os"
	"fmt"
	"path"
	"strings"
)

type Reference struct{
	Name, Url	string
	Checksum	[32]byte
}

func (reference Reference) list(){
	fmt.Printf("Name:\t\t%v\nUrl:\t\t%v\nChecksum:\t%v\n", reference.Name, reference.Url, strings.Trim(fmt.Sprintf("%v", reference.Checksum), "[]"))
}

type References []Reference

func (references References) write(file string){
	_, err := os.Stat(path.Dir(file))
	if os.IsNotExist(err){
		fmt.Printf("FAIL!\nDirectory %v does not exist!\n\n", path.Dir(file))
		panic(err)
	}

	data, _ := json.Marshal(references)

	err = ioutil.WriteFile(file, data, 0666)
	if err != nil{
		fmt.Printf("FAIL!\nCould not write to %v\n\n", file)
		panic(err)
	}
}

func (references References) find(file string) int{
	for i, reference := range references{
		if reference.Name == path.Base(file){
			return i
		}
	}

	fmt.Printf("FAIL!\nNo reference %v in references!\n\n", path.Base(file))
	panic(nil)
}

func (references References) validate(file string, codes []int) Reference{
	comparison := Reference{
		Name:		path.Base(file),
		Url:		url(file),
		Checksum:	checksum(file),
	}

	black := false
	zero := true

	messages:= []string{
		"No similar reference found!",
		"Found reference with same content(SHA256)!",
		"Something went horribly wrong!",
		"Something went horribly wrong!",
		"Reference name already assigned!",
		"Found same reference with different URL!",
		"Found same reference with differenct content(SHA256)!",
		"Found identical reference!"}

	for _, reference := range references{
		value := 0

		if reference.Name == comparison.Name{
			value += 4
		}
		if reference.Url == comparison.Url{
			value += 2
		}
		if reference.Checksum == comparison.Checksum{
			value += 1
		}

		for i, code := range codes{
			if value == code{
				zero = false
				break
			}

			if i == len(codes)-1 && value > 0{
				fmt.Printf("FAIL!\n" + messages[value] + "\nCurrent:\n\n")
				comparison.list()
				fmt.Printf("\nSimilar:\n\n")
				reference.list()
				fmt.Printf("\n")

				black = true
			}
		}
	}

	if black{
		panic(nil)
	}
	if len(references) > 0 && zero{
		fmt.Printf("FAIL!\n" + messages[0] + "\n\n")
		panic(nil)
	}

	return comparison
}

func main(){
	if len(os.Args) > 1{
		switch os.Args[1]{
		case "new":
			if len(os.Args) == 3{
				newreferences(os.Args[2])
			} else {
				help()
			}
		case "add":
			if len(os.Args) == 4{
				addreference(os.Args[2], os.Args[3])
			} else {
				help()
			}
		case "remove":
			if len(os.Args) == 4{
				removereference(os.Args[2], os.Args[3])
			} else {
				help()
			}
		case "update":
			if len(os.Args) == 4{
				updatereference(os.Args[2], os.Args[3])
			} else {
				help()
			}
		case "list":
			if len(os.Args) == 3{
				listreferences(os.Args[2])
			} else {
				help()
			}
		default:
			help()
		}
	} else{
		help()
	}
}

func newreferences(name string){
	fmt.Printf("Creating new references... ")

	var references References
	references.write(name)

	fmt.Printf("Done!\n")
}

func addreference(name, file string){
	fmt.Printf("Adding reference to references... ")

	references := read(name)

	reference := references.validate(file, []int{0})

	references = append(references, reference)

	references.write(name)

	fmt.Printf("Done!\n")
}

func removereference(name, reference string){
	fmt.Printf("Removing reference from references... ")

	references := read(name);

	i := references.find(reference)

	references[i]=references[len(references)-1]

	references[:len(references)-1].write(name)

	fmt.Printf("Done!\n")
}

func updatereference(name, file string){
	fmt.Printf("Updating reference in references... ")

	references := read(name)

	reference := references.validate(file, []int{4, 5, 6})

	i := references.find(file)

	references[i] = reference

	references.write(name)

	fmt.Printf("Done!\n")
}

func listreferences(name string){
	references := read(name)

	if len(references) < 1{
		fmt.Printf("No reference found in references %v!\n", name)
		panic(nil)
	}

	references[0].list()

	for _, reference := range references[1:]{
		fmt.Printf("\n")
		reference.list()
	}
}

func help(){
	fmt.Printf("Usage: ref COMMAND [ARGUMENT]\nCreate references and add, remove or modify a reference.\n\nCOMMANDS:\n\tnew NAME\t\tCreate new refernces named NAME.\n\tadd NAME FILE\t\tAdd reference FILE to references NAME.\n\tremove NAME REFERENCE\tRemove REFERENCE from reference NAME.\n\tupdate NAME FILE\tUpdate reference FILE in reference NAME.\n\tlist NAME\t\tList references NAME.\n\nARGUMENTS:\n\tNAME\t\tValid JSON file name.\n\tFILE\t\tAbsolute Path to file.\n\tREFERENCE\tName of a reference in references.\n")
}

func read(file string) References{
	data, err := ioutil.ReadFile(file)
	if err != nil{
		fmt.Printf("FAIL!\nCould not read from %v!\n\n", file)
		panic(err)
	}

	var references References
	err = json.Unmarshal(data, &references)
	if err != nil{
		fmt.Printf("FAIL!\n%v is not a JSON file!\n\n", file)
		panic(err)
	}

	return references
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
	url := split[len(split) - 1]

	repository := strings.TrimSuffix(path.Base(url), path.Ext(url))
	url = path.Dir(url)
	repository = path.Join(path.Base(url), repository)

	return path.Join("https://api.github.com/repos", repository, "contents", content + "?ref=master")
}

func checksum(file string) [32]byte{
	data, err := ioutil.ReadFile(file)
	if err != nil{
		fmt.Printf("FAIL!\nCould not read from %v!\n\n", file)
		panic(err)
	}

	return sha256.Sum256(data)
}
