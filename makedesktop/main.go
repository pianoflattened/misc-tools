package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
	"github.com/pborman/getopt/v2"
)

const EXEC_PAGE_LIMIT = 9

var (
	description=""
	resolution=""
	gamedir=""
	help bool
)

func init() {
	getopt.FlagLong(&description, "description", 'd', "write the description. keep it short")
	getopt.FlagLong(&resolution, "res", 'r', "force a resolution")
	getopt.FlagLong(&help, "help", 'h', "display this text and exit")
}

func main() {
	getopt.Parse()

	if help {
		getopt.Usage()
		return
	}

	fmt.Println(description)

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	args := getopt.Args()
	if len(args) == 0 {
		fmt.Printf("no exectuable specified, checking current directory (%s)..\n", wd)
		executable := find_executable()
		if executable == "" {
			fmt.Println("none found")
			return
		}
	}

	if len(args) > 1 {
		fmt.Println("why did u use more than 1 arg im only taking the first one get fucked LOL")
	}

	exe := filepath.Base(args[0])
	name := filepath.Base(filepath.Dir(args[0]))
	if name == "." {
		name = filepath.Base(wd)
	}
	var exe_abs string; exe_abs, err = filepath.Abs(args[0]); if err != nil { panic(err) }
	exedir_abs := filepath.Dir(exe_abs)
	gamedir := strings.ReplaceAll(strings.ToLower(name), " ", "-")

	err = exec.Command("sh", "-c", strings.Join([]string{"cp -r", strings.ReplaceAll(sh_escape(strings.TrimSuffix(exedir_abs, "/")+"/*"), " ", "\\ "), "$HOME/.local/games/"+sh_escape(gamedir)}, " ")).Run(); if err != nil { panic(err) }
	
	desktop_contents := format_desktop_file(name, description, gamedir, exe, resolution)
	err = exec.Command("sh", "-c", strings.Join([]string{"echo -e", desktop_contents, ">>", "$HOME/.local/share/applications/"+sh_escape(gamedir)+".desktop"}, " ")).Run(); if err != nil { panic(err) }
	fmt.Println("done!")
}

func find_executable() string {
	executables := []string{}
	
	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		var abs string; abs, err = filepath.Abs(info.Name())
		if err != nil {
			panic(err)
		}
		
		if is_executable(abs) {
			executables = append(executables, info.Name())
		}
		
		return err
	})
	if err != nil { panic(err) }
	
	switch (len(executables)) {
	case 0:
		return ""
	case 1:
		return executables[0]
	default:
		return make_exec_choose_menu(executables, 1)
	}
	
	return ""
}

func make_exec_choose_menu(executables []string, page int) string {
	first_index := EXEC_PAGE_LIMIT*(page-1)
	last_index := len(executables)-1
	last_page := int(math.Ceil(float64(len(executables)) / float64(EXEC_PAGE_LIMIT)))
	padding := len(fmt.Sprintf("%d", last_index))

	choosestr := ""
	for index, executable := range executables {
		if index < first_index {
			continue
		}
	
		if index-((page-1)*EXEC_PAGE_LIMIT) > EXEC_PAGE_LIMIT {
			choosestr += "(n)ext"
			
			if page > 1 {
				choosestr += "\t(p)revious\n"
			} else {
				choosestr += "\n"
			}
			
			break
		}
	
		choosestr += fmt.Sprintf("%0*d) %s\n", index, padding, executable)

		if index == last_index {
			choosestr += "(p)revious\n"
		}
	}

	fmt.Printf(choosestr)
	fmt.Print("> ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	input = strings.TrimSuffix(input, "\n")

	if input == "" {
		return executables[0]
	} else if _, err := strconv.Atoi(input); err == nil {
		n, _ := strconv.Atoi(input)

		if n < 0 || n > EXEC_PAGE_LIMIT {
			return make_exec_choose_menu(executables, page)
		} else {
			return executables[n]
		}
	} else if input == "n" {
		if page == last_page {
			return make_exec_choose_menu(executables, page)
		} else {
			return make_exec_choose_menu(executables, page+1)
		}
	} else if input == "p" {
		if page == 1 {
			return make_exec_choose_menu(executables, page)
		} else {
			return make_exec_choose_menu(executables, page-1)
		}
	} else {
		return make_exec_choose_menu(executables, page)
	}
}

func sh_escape(str string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(str, "\\", "\\\\"), "\"", "\\\""), "'", "\\'")
}

func format_desktop_file(name, description, gamedir, executable, resolution string) string {
	res_str := format_res_string(resolution)
	return fmt.Sprintf(`"[Desktop Entry]\nType=Application\nVersion=1.0\nName=`+sh_escape(name)+`\nComment=`+sh_escape(description)+`\nExec=sh -c \"cd $HOME/.local/games/`+sh_escape(gamedir)+` && wine`+sh_escape(res_str)+` `+sh_escape(executable)+`\"\nIcon=wine\nTerminal=false\nCategories=Games;"`)
}

func format_res_string(resolution string) (n string) {
	if len(resolution) > 0 {
		n += "vd"

		if resolution != "640x480" {
			n += " -r "+resolution
		}
	} 
	
	return
}

func is_executable(file string) bool {
	fileInfo, err := os.Stat(file)
	if err != nil{
		return false
	}
	
	m := fileInfo.Mode()    
	if !( (m.IsRegular()) || (uint32(m & fs.ModeSymlink) == 0) ) {
		return false
	}
	if (uint32(m & 0111) == 0){
        return false
    }
    
	if unix.Access(file, unix.X_OK) != nil {
		return false
	}
	
	return true
}
