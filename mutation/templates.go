package mutation

import (
	"bruce/loader"
	"github.com/rs/zerolog/log"
	"text/template"
)

func WriteInlineTemplate(filename, tpl string, content interface{}) error {
	t, err := template.New("generator").Parse(tpl)
	if err != nil {
		log.Error().Err(err).Msg("could not parse cron template")
		return err
	}
	w, err := loader.WriterFromLocal(filename)
	if err != nil {
		return err
	}
	defer w.Close()
	return t.Execute(w, content)
}
