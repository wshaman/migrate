package migrate

import "testing"

func Test_cliExec_STDERR(t *testing.T) {
	_, err := cliExec("CommMand_SHOUlD_NoT_Exist")
	if err == nil {
		t.Errorf("should be an error here")
	}
}
