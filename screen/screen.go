package screen

import (
	"fmt"
	"image"
	"log"
	"runtime"
	"strings"

	"github.com/funsun/peridot/common"
	"github.com/funsun/peridot/controller"

	"github.com/go-gl/gl/v4.5-compatibility/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const vertexShaderSource = `
	#version 410 core

	layout (location = 0) in vec3 position;
	layout (location = 1) in vec3 color;
	layout (location = 2) in vec2 texCoord;

	out vec3 ourColor;
	out vec2 TexCoord;

	void main()
	{
	    gl_Position = vec4(position, 1.0);
	    ourColor = color;       // pass the color on to the fragment shader
	    TexCoord = texCoord;    // pass the texture coords on to the fragment shader
	}
`

const fragmentShaderSource = `
	#version 410 core

	in vec3 ourColor;
	in vec2 TexCoord;

	out vec4 color;

	uniform sampler2D ourTexture0;
	uniform sampler2D ourTexture1;

	void main()
	{
	    // mix the two textures together (texture1 is colored with "ourColor")
	    color = mix(texture(ourTexture0, TexCoord), texture(ourTexture1, TexCoord) * vec4(ourColor, 1.0f), 0.5);
	}
`

type getObjIv func(uint32, uint32, *int32)
type getObjInfoLog func(uint32, int32, *int32, *uint8)

func getGlError(glHandle uint32, checkTrueParam uint32, getObjIvFn getObjIv,
	getObjInfoLogFn getObjInfoLog, failMsg string) error {

	var success int32
	getObjIvFn(glHandle, checkTrueParam, &success)

	if success == gl.FALSE {
		var logLength int32
		getObjIvFn(glHandle, gl.INFO_LOG_LENGTH, &logLength)

		log := gl.Str(strings.Repeat("\x00", int(logLength)))
		getObjInfoLogFn(glHandle, logLength, nil, log)

		return fmt.Errorf("%s: %s", failMsg, gl.GoStr(log))
	}

	return nil
}

var vertices = []float32{
	// top left
	-0.75, 0.75, 0.0, // position
	1.0, 1.0, 1.0,
	0.0, 0.0,

	// top right
	0.75, 0.75, 0.0,
	1.0, 1.0, 1.0,
	1.0, 0.0, // texture coordinates

	// bottom right
	0.75, -0.75, 0.0,
	1.0, 1.0, 1.0,
	1.0, 1.0,

	// bottom left
	-0.75, -0.75, 0.0,
	1.0, 1.0, 1.0,
	0.0, 1.0,
}

var indices = []uint32{
	// rectangle
	0, 1, 2, // top triangle
	3, 0, 2, // bottom triangle
}

type Screen struct {
	w, h   int
	handle *glfw.Window
	prog   *Program
	tex    *Texture
	img    image.Image
	ctrl   *controller.Controller
	vao    *VAO
	Done   chan bool
}

func (s *Screen) Init(w, h int, c *controller.Controller) *Screen {
	s.w, s.h = w, h
	s.ctrl = c
	s.Done = make(chan bool)
	return s
}

func (s *Screen) Show() {
	go func() {
		runtime.LockOSThread()
		defer glfw.Terminate()

		s.initWindow()
		s.initOpenGL()
		s.vao = new(VAO).Init(vertices, indices)
		s.beforeLoop()
		for !s.handle.ShouldClose() {
			s.onUpdate()
		}
		common.Terminate <- true
	}()
}

func (s *Screen) AddFrameBuffer(img image.Image) {
	s.img = img
}

func (s *Screen) initWindow() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	var err error
	s.handle, err = glfw.CreateWindow(s.w, s.h, "FunSun's Peridot", nil, nil)
	// bit:   	 7     6     5     4     3     2     1     0
	// button:	 A     B  Select Start  Up   Down  Left  Right
	s.handle.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		k := uint8(0)
		switch key {
		case glfw.KeyRight:
			k = controller.Right
		case glfw.KeyLeft:
			k = controller.Left
		case glfw.KeyDown:
			k = controller.Down
		case glfw.KeyUp:
			k = controller.Up
		case glfw.KeyX:
			k = controller.Start
		case glfw.KeyZ:
			k = controller.Select
		case glfw.KeyS:
			k = controller.B
		case glfw.KeyA:
			k = controller.A
		default:
			return
		}
		if action == glfw.Press {
			s.ctrl.SetButton(k)
			// 看来action不知有Press和Release， 持续按下产生的action不是这两种之一
		} else if action == glfw.Release {
			s.ctrl.ClearButton(k)
		}
	})
	if err != nil {
		panic(err)
	}
	s.handle.MakeContextCurrent()
}

func (s *Screen) initOpenGL() {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

}

func (s *Screen) beforeLoop() {
	var err error
	v, err := NewShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err.Error())
	}
	f, err := NewShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err.Error())
	}

	s.prog, _ = NewProgram(v, f)
	if err != nil {
		panic(err.Error())
	}

	s.prog.Use()

}

func (s *Screen) onUpdate() {
	defer s.handle.SwapBuffers()
	glfw.PollEvents()
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	if s.img != nil {
		s.tex, _ = NewTexture(s.img, gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
		s.img = nil
	}
	if s.tex == nil {
		return
	}
	s.tex.Bind(gl.TEXTURE0)
	s.tex.SetUniform(s.prog.GetUniformLocation("ourTexture0"))
	s.vao.Bind()
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, gl.PtrOffset(0))
	s.vao.Unbind()
	s.tex.UnBind()
}
