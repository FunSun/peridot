package screen

import "github.com/go-gl/gl/v4.5-compatibility/gl"

type Program struct {
	handle  uint32
	shaders []*Shader
}

func (prog *Program) Delete() {
	for _, shader := range prog.shaders {
		shader.Delete()
	}
	gl.DeleteProgram(prog.handle)
}

func (prog *Program) Attach(shaders ...*Shader) {
	for _, shader := range shaders {
		gl.AttachShader(prog.handle, shader.handle)
		prog.shaders = append(prog.shaders, shader)
	}
}

func (prog *Program) Use() {
	gl.UseProgram(prog.handle)
}

func (prog *Program) Link() error {
	gl.LinkProgram(prog.handle)
	return getGlError(prog.handle, gl.LINK_STATUS, gl.GetProgramiv, gl.GetProgramInfoLog,
		"PROGRAM::LINKING_FAILURE")
}

func (prog *Program) GetUniformLocation(name string) int32 {
	return gl.GetUniformLocation(prog.handle, gl.Str(name+"\x00"))
}

func NewProgram(shaders ...*Shader) (*Program, error) {
	prog := &Program{handle: gl.CreateProgram()}
	prog.Attach(shaders...)

	if err := prog.Link(); err != nil {
		return nil, err
	}

	return prog, nil
}
