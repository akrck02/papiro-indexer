package command

import (
	"os"

	"github.com/akrck02/papiro-indexer/logger"
	"github.com/joho/godotenv"
)

func LoadEnv(filePath string) {
	err := godotenv.Load(filePath)

	if nil != err {
		println("Error loading .env file")
	}

	checkCompulsoryVariables()
}

func checkCompulsoryVariables() {
	logger.Log(" 📚", "Output path:", os.Getenv("WIKI_PATH"), "📚")
	logger.Log()

}
