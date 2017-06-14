package screen

import (
	"github.com/go-gl/gl/v4.5-compatibility/gl"
	"io/ioutil"
)

type Shader struct {
	handle uint32
}

func (shader *Shader) Delete() {
	gl.DeleteShader(shader.handle)
}

func NewShader(src string, sType uint32) (*Shader, error) {

	handle := gl.CreateShader(sType)
	glSrc, free := gl.Strs(src + "\x00")
	gl.ShaderSource(handle, 1, glSrc, nil)
	free()
	gl.CompileShader(handle)
	err := getGlError(handle, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog,
		"SHADER::COMPILE_FAILURE::")
	if err != nil {
		return nil, err
	}
	return &Shader{handle: handle}, nil
}

func NewShaderFromFile(file string, sType uint32) (*Shader, error) {
	src, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	handle := gl.CreateShader(sType)
	glSrc := gl.Str(string(src) + "\x00")
	gl.ShaderSource(handle, 1, &glSrc, nil)
	gl.CompileShader(handle)
	err = getGlError(handle, gl.COMPILE_STATUS, gl.GetShaderiv, gl.GetShaderInfoLog,
		"SHADER::COMPILE_FAILURE::"+file)
	if err != nil {
		return nil, err
	}
	return &Shader{handle: handle}, nil
}
