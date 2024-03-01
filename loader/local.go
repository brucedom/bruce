package loader

import (
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"path"
)

func ReadFromLocal(fileName string) ([]byte, string, error) {
	fn := path.Base(fileName)
	log.Debug().Msgf("starting local read of %s", fileName)
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		log.Info().Msgf("local reader engine: (file does not exist): %s", fileName)
		return nil, fn, err
	}
	f, err := os.Open(fileName)
	if err != nil {
		log.Error().Err(err).Msgf("could not open file for reading: %s", fileName)
		return nil, fn, err
	}
	d, err := io.ReadAll(f)
	if err != nil {
		log.Error().Err(err).Msgf("could not read local file: %s", fileName)
		return nil, fn, err
	}
	log.Error().Err(f.Close())
	return d, fn, err
}

func WriterFromLocal(fileName string) (io.WriteCloser, error) {
	w, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		log.Error().Err(err).Msgf("could not open file for writing: %s", fileName)
		return nil, err
	}
	return w, nil
}
