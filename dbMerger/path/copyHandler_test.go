package path

import (
	"os"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCopyHandler(t *testing.T) {
	t.Parallel()

	handler := NewCopyHandler()
	assert.False(t, check.IfNil(handler))
}

func TestCopyHandler_CopyDirectory(t *testing.T) {
	t.Parallel()

	err := os.RemoveAll("./testdata/destDir")
	assert.Nil(t, err)

	handler := NewCopyHandler()
	err = handler.CopyDirectory("./testdata/destDir", "./testdata/srcDir")
	assert.Nil(t, err)

	assert.Equal(t, readFileContent(t, "./testdata/srcDir/a/1"), readFileContent(t, "./testdata/destDir/a/1"))
	assert.Equal(t, readFileContent(t, "./testdata/srcDir/a/file2.file"), readFileContent(t, "./testdata/destDir/a/file2.file"))
	assert.Equal(t, readFileContent(t, "./testdata/srcDir/b/a.txt"), readFileContent(t, "./testdata/destDir/b/a.txt"))
	assert.Equal(t, readFileContent(t, "./testdata/srcDir/c.log"), readFileContent(t, "./testdata/destDir/c.log"))
}

func readFileContent(tb testing.TB, path string) string {
	in, err := os.Open(path)
	require.Nil(tb, err)

	defer func() {
		errClose := in.Close()
		require.Nil(tb, errClose)
	}()

	buff := make([]byte, 10000)
	n, err := in.Read(buff)
	require.Nil(tb, err)

	contents := string(buff[:n])
	log.Info("read file", "path", path, "contents", contents)

	return contents
}
