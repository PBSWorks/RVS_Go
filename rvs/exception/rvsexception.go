package exception

type RVSError struct {
	Errordetails string
	Errorcode    string
	Errortype    string
}

func (e *RVSError) Error() string {
	return e.Errordetails
}
