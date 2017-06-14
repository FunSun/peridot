package screen

import "github.com/go-gl/gl/v4.5-compatibility/gl"

type VAO struct {
	handle uint32
	vbo    *VBO
	ebo    *EBO
}

func (v *VAO) Init(vertices []float32, indices []uint32) *VAO {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	v.handle = vao

	v.vbo = new(VBO).Init(vertices, gl.STATIC_DRAW)
	v.ebo = new(EBO).Init(indices, gl.STATIC_DRAW)

	// size of one whole vertex (sum of attrib sizes)
	var stride int32 = 3*4 + 3*4 + 2*4
	var offset int = 0

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(0)
	offset += 3 * 4

	// color
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(1)
	offset += 3 * 4

	// texture position
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(offset))
	gl.EnableVertexAttribArray(2)
	offset += 2 * 4

	// unbind the VAO (safe practice so we don't accidentally (mis)configure it later)
	gl.BindVertexArray(0)

	return v
}

func (v *VAO) Bind() {
	gl.BindVertexArray(v.handle)
}

func (v *VAO) Unbind() {
	gl.BindVertexArray(0)
}

type VBO struct {
	handle uint32
}

// typ: gl. gl.STATIC_DRAW etc
func (v *VBO) Init(vertices []float32, typ uint32) *VBO {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), typ)
	v.handle = vbo
	return v
}

type EBO struct {
	handle uint32
}

func (e *EBO) Init(indices []uint32, typ uint32) *EBO {
	var ebo uint32
	gl.GenBuffers(1, &ebo)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), typ)
	e.handle = ebo
	return e
}
