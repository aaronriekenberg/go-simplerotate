package rotation

import (
	"fmt"
	"os"

	"github.com/aaronriekenberg/go-simplerotate/constants"
	"github.com/aaronriekenberg/go-simplerotate/logging"
)

var logger = logging.GetLogger()

type rotationAction struct {
	fromFileName string
	toFileName   string
}

func (rotationAction rotationAction) rotate() {
	logger.Printf("rotate from %q to %q", rotationAction.fromFileName, rotationAction.toFileName)
	os.Rename(rotationAction.fromFileName, rotationAction.toFileName)
}

func buildRotationActions() []rotationAction {
	if constants.MaxOutputFiles <= 1 {
		return nil
	}

	buildFileName := func(fileIndex int) string {
		switch fileIndex {
		case 0:
			return constants.OutputFileName
		default:
			return fmt.Sprintf("%v.%v", constants.OutputFileName, fileIndex)
		}
	}

	rotationActions := make([]rotationAction, 0, constants.MaxOutputFiles-1)
	for i := constants.MaxOutputFiles - 1; i > 0; i-- {
		fromFileName := buildFileName(i - 1)
		toFileName := buildFileName(i)
		rotationActions = append(rotationActions,
			rotationAction{
				fromFileName: fromFileName,
				toFileName:   toFileName,
			},
		)
	}

	return rotationActions
}

var rotationActions []rotationAction

func init() {
	rotationActions = buildRotationActions()

	logger.Printf("rotationActions = %+v", rotationActions)
}

func RotateOutputFiles() {
	for _, action := range rotationActions {
		action.rotate()
	}
}
