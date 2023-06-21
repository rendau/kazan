package errs

import (
	"github.com/rendau/dop/dopErrs"
)

const (
	BadFormData = dopErrs.Err("bad_form_data")
	BadFile     = dopErrs.Err("bad_file")
)
