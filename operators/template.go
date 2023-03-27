package operators

import (
	"bruce/exe"
	"bruce/loader"
	"bruce/random"
	"bruce/system"
	"bytes"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
)

var (
	// may want to re-use this later but tbd
	templateFuncs = template.FuncMap{
		"contains": strings.Contains,
		"dump":     func(field interface{}) string { return dump(field) },
	}
	backupDir string
)

func init() {
	backupDir = fmt.Sprintf("%s%c%s", os.TempDir(), os.PathSeparator, random.String(12))
	log.Info().Msgf("using template backup directory as: %s", backupDir)
	os.MkdirAll(backupDir, 0775)
}

type Template struct {
	Template  string      `yaml:"template"`
	RemoteLoc string      `yaml:"remoteLocation"`
	Perms     os.FileMode `yaml:"perms"`
	Owner     string      `yaml:"owner"`
	Group     string      `yaml:"group"`
	Variables []TVars     `yaml:"vars"`
}
type TVars struct {
	ObType   string `yaml:"type"`
	Input    string `yaml:"input"`
	Variable string `yaml:"variable"`
}

func dump(field interface{}) string {
	buf := &bytes.Buffer{}
	spew.Fdump(buf, field)
	return buf.String()
}

func (t *Template) Execute() error {
	// backup existing template if exists
	if exe.FileExists(t.Template) {
		log.Debug().Msgf("backing up existing template: %s", t.Template)
		err := exe.CopyFile(t.Template, fmt.Sprintf("%s%c%s", backupDir, os.PathSeparator, path.Base(t.Template)), false)
		if err != nil {
			log.Debug().Err(err).Msgf("could not create backup file: %s", err)
			return err
		}
	} else {
		log.Debug().Str("template", t.Template).Msg("no existing template file exists")
	}
	return ExecuteTemplate(t.Template, t.RemoteLoc, t.Variables, t.Perms)
	// run template exec on file
}

func GetBackupFileChecksum(src string) (string, error) {
	backupFileName := fmt.Sprintf("%s%c%s", backupDir, os.PathSeparator, strings.TrimLeft(src, string(os.PathSeparator)))
	return exe.GetFileChecksum(backupFileName)
}

func ExecuteTemplate(local, remote string, vars []TVars, perms fs.FileMode) error {
	// we have the backup so now we can delete the file if it exists
	if exe.FileExists(local) {
		exe.DeleteFile(local)
	} else {
		// check if the directories exist to render the file
		if !exe.FileExists(path.Dir(local)) {
			os.MkdirAll(path.Dir(local), 0775)
		}
	}

	log.Debug().Msgf("template exec starting on: %s", local)
	t, err := loadTemplateFromRemote(remote)
	if err != nil {
		log.Err(err).Msgf("cannot read template source %s", local)
		return err
	}
	var content = make(map[string]string)
	for _, v := range vars {
		content[v.Variable] = loadTemplateValue(v)
	}

	destination, err := os.OpenFile(local, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		log.Error().Err(err).Msgf("could not open file for writing template: %s", local)
		return err
	}
	defer destination.Close()
	err = t.Execute(destination, content)
	if err != nil {
		log.Err(err).Msgf("could not write template: %s", local)
		return err
	}
	log.Info().Msgf("template written: %s", local)
	localHash, err := exe.GetFileChecksum(local)
	if err != nil {
		log.Error().Err(err).Msg("failed to get new file checksum")
	}
	backupHash, err := GetBackupFileChecksum(local)
	if err != nil {
		// no backup exists so lets add it as a changed template as it should be net new.
		log.Info().Msgf("no backup file for %s", local)
	}
	if localHash != backupHash {
		system.Get().AddModifiedTemplate(local)
	}
	return nil
}

func loadTemplateValue(v TVars) string {
	if v.ObType == "value" {
		return GetValueForOSHandler(v.Input)
	}
	if v.ObType == "command" {
		var outb, errb bytes.Buffer
		cText := strings.Fields(v.Input)
		if len(cText) > 1 {
			cmd := exec.Command(cText[0], cText[1:]...)
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			err := cmd.Run()
			if err != nil {
				log.Err(err).Msg("error executing command returning error statement")
				// we don't want to put crazy errors in our templates anyway...
				return "ERROR_IN_CMD"
			}
		} else {
			cmd := exec.Command(v.Input)
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			err := cmd.Run()
			if err != nil {
				log.Err(err).Msg("error executing command returning error statement")
				// we don't want to put crazy errors in our templates anyway...
				return "ERROR_IN_CMD"
			}
		}
		fmt.Println(cText[0])

		return outb.String()
	}
	// sometimes we will actually want an empty string so this is okay
	return ""
}

func loadTemplateFromRemote(remoteLoc string) (*template.Template, error) {
	d, _, err := loader.ReadRemoteFile(remoteLoc)
	if err != nil {
		log.Error().Err(err).Msgf("could not read remote template file: %s", remoteLoc)
	}
	log.Debug().Msgf("remote template read completed for: %s", remoteLoc)
	return template.New(path.Base(remoteLoc)).Parse(string(d))
}
