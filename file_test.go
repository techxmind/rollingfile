package rollingfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RollingFileTestSuite struct {
	suite.Suite
	tmpdir          string
	dftFilename     string
	dftRotatedName  string
	dftRotatedName2 string
}

func (s *RollingFileTestSuite) SetupTest() {
	currentTime = func() time.Time {
		t, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
		return t
	}

	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(fmt.Sprintf("unexpected ioutil.TempDir error: %v", err))
	}
	s.tmpdir = tmpdir

	s.dftFilename = filepath.Join(tmpdir, "base.json")
	s.dftRotatedName = filepath.Join(tmpdir, "base-20060102150405.json")
	s.dftRotatedName2 = filepath.Join(tmpdir, "base-20060102150405.1.json")
}

func (s *RollingFileTestSuite) TearDownTest() {
	os.RemoveAll(s.tmpdir)
	currentTime = time.Now
}

func (s *RollingFileTestSuite) TestRotatedName() {
	f, _ := New("base.json")

	s.Equal("base-20060102150405.json", f.rotatedName())

	f.suffix = "-test"
	s.Equal("base-20060102150405-test.json", f.rotatedName())
}

func (s *RollingFileTestSuite) TestOpenNew() {
	t := s.T()
	filename := filepath.Join(s.tmpdir, "data", "base.json")
	f, err := New(filename)
	require.NoError(t, err)

	err = f.openNew()
	require.NoError(t, err)

	assert.FileExists(t, filename)
	assert.NotNil(t, f.current)
	f.close()
	assert.Nil(t, f.current)

	err = f.openNew()
	require.NoError(t, err)
	assert.NotNil(t, f.current)
}

func (s *RollingFileTestSuite) TestRotate() {
	t := s.T()
	filename := filepath.Join(s.tmpdir, "data", "base.json")
	os.Mkdir(filepath.Join(s.tmpdir, "data"), 0700)

	err := ioutil.WriteFile(filename, []byte("hello"), os.FileMode(0600))
	require.NoError(t, err)

	f, err := New(filename)
	require.NoError(t, err)

	f.rotate()

	rotatedFile := filepath.Join(s.tmpdir, "data", "base-20060102150405.json")
	require.FileExists(t, rotatedFile)
	require.FileExists(t, filename)

	stat, _ := os.Stat(rotatedFile)
	assert.Equal(t, int64(5), stat.Size())

	stat, _ = os.Stat(filename)
	assert.Equal(t, int64(0), stat.Size())

	assert.NotNil(t, f.current)
}

func (s *RollingFileTestSuite) MustGetInstance(filename string, opts ...Option) *RollingFile {
	f, err := New(filename, opts...)
	require.NoError(s.T(), err)
	return f
}

func (s *RollingFileTestSuite) TestOpenCurrent() {
	f := s.MustGetInstance(s.dftFilename, MaxSize(10))
	f.openCurrent(5)
	s.FileExists(s.dftFilename)
	s.NoFileExists(s.dftRotatedName)
	f.close()
	os.Remove(s.dftFilename)
	s.NoFileExists(s.dftFilename)
	ioutil.WriteFile(s.dftFilename, []byte("hello"), 0644)
	f.openCurrent(4)
	s.FileExists(s.dftFilename)
	s.NoFileExists(s.dftRotatedName)
	f.close()
	os.Remove(s.dftFilename)
	ioutil.WriteFile(s.dftFilename, []byte("hello"), 0644)
	f.openCurrent(5)
	s.FileExists(s.dftFilename)
	s.FileExists(s.dftRotatedName)
	s.NotNil(f.current)
}

func (s *RollingFileTestSuite) TestWrite() {
	f := s.MustGetInstance(s.dftFilename, MaxSize(10), MaxAge(60))
	n, err := f.Write([]byte("hello"))
	s.NoError(err)
	s.Equal(5, n)
	s.NoFileExists(s.dftRotatedName)
	n, err = f.Write([]byte("word"))
	s.NoError(err)
	s.Equal(4, n)
	s.NoFileExists(s.dftRotatedName)
	n, err = f.Write([]byte("!"))
	s.NoError(err)
	s.Equal(1, n)
	s.FileExists(s.dftRotatedName)
	p, _ := ioutil.ReadFile(s.dftRotatedName)
	s.Equal("helloword!", string(p))
	n, err = f.Write([]byte("hello"))
	s.NoError(err)
	s.Equal(5, n)
	currentTime = func() time.Time {
		t, _ := time.Parse(time.RFC3339, "2006-01-02T15:05:06Z")
		return t
	}
	n, err = f.Write([]byte("world"))
	s.NoError(err)
	s.Equal(5, n)
	rotatedName := filepath.Join(s.tmpdir, "base-20060102150506.json")
	s.FileExists(rotatedName)
	p, _ = ioutil.ReadFile(rotatedName)
	s.Equal("hello", string(p))
	n, err = f.Write([]byte("worldworld"))
	s.NoError(err)
	s.Equal(10, n)
	rotatedName = filepath.Join(s.tmpdir, "base-20060102150506.1.json")
	s.FileExists(rotatedName)
	p, _ = ioutil.ReadFile(rotatedName)
	s.Equal("worldworldworld", string(p))
}

func TestRollingFileTestSuite(t *testing.T) {
	suite.Run(t, new(RollingFileTestSuite))
}
