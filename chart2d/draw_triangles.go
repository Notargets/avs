package chart2d

//func init() {
//	runtime.LockOSThread()
//}

//func (cc *Chart2D) Init() {
//	vertices := []float32{
//		-0.5, -0.5, 0.0, 1.0, 0.0, 0.0, 1.0,
//		0.5, -0.5, 0.0, 0.0, 1.0, 0.0, 1.0,
//		0.0, 0.5, 0.0, 0.0, 0.0, 1.0, 1.0,
//	}
//
//	gl.GenVertexArrays(1, &cc.VAO)
//	gl.GenBuffers(1, &cc.VBO)
//
//	gl.BindVertexArray(cc.VAO)
//	gl.BindBuffer(gl.ARRAY_BUFFER, cc.VBO)
//	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
//
//	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 7*4, unsafe.Pointer(uintptr(0)))
//	gl.EnableVertexAttribArray(0)
//
//	gl.VertexAttribPointer(1, 4, gl.FLOAT, false, 7*4, unsafe.Pointer(uintptr(3*4)))
//	gl.EnableVertexAttribArray(1)
//
//	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
//	gl.BindVertexArray(0)
//
//	cc.Shader = cc.compileShaders()
//}

//func (cc *Chart2D) compileShaders() uint32 {
//	vertexShaderSource := `
//	#version 450
//	layout (location = 0) in vec3 position;
//	layout (location = 1) in vec4 color;
//	uniform mat4 MVP;
//	out vec4 fragColor;
//	void main() {
//		fragColor = color;
//		gl_Position = MVP * vec4(position, 1.0);
//	}
//	` + "\x00"
//
//	fragmentShaderSource := `
//	#version 450
//	in vec4 fragColor;
//	out vec4 color;
//	void main() {
//		color = fragColor;
//	}
//	` + "\x00"
//
//	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
//	cVertexShaderSource, freeVertexShaderSource := gl.Strs(vertexShaderSource)
//	defer freeVertexShaderSource() // Free memory after compilation
//	gl.ShaderSource(vertexShader, 1, cVertexShaderSource, nil)
//	gl.CompileShader(vertexShader)
//
//	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
//	cFragmentShaderSource, freeFragmentShaderSource := gl.Strs(fragmentShaderSource)
//	defer freeFragmentShaderSource() // Free memory after compilation
//	gl.ShaderSource(fragmentShader, 1, cFragmentShaderSource, nil)
//	gl.CompileShader(fragmentShader)
//
//	shaderProgram := gl.CreateProgram()
//	gl.AttachShader(shaderProgram, vertexShader)
//	gl.AttachShader(shaderProgram, fragmentShader)
//	gl.LinkProgram(shaderProgram)
//
//	gl.DeleteShader(vertexShader)
//	gl.DeleteShader(fragmentShader)
//
//	return shaderProgram
//}
//
//func (cc *Chart2D) Render() {
//	// 1. Clear the screen
//	gl.ClearColor(0.1, 0.1, 0.1, 1.0) // Background color
//	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
//
//	// 2. Calculate MVP matrix (Projection * View * Model)
//	projection := mgl32.Ortho2D(-1, 1, -1, 1)
//	model := mgl32.Translate3D(cc.Position[0], cc.Position[1], 0).Mul4(mgl32.Scale3D(cc.Scale, cc.Scale, 1))
//	mvp := projection.Mul4(model)
//
//	// 3. Pass the MVP to the shader
//	gl.UseProgram(cc.Shader)
//	mvpUniform := gl.GetUniformLocation(cc.Shader, gl.Str("MVP\x00"))
//	gl.UniformMatrix4fv(mvpUniform, 1, false, &mvp[0])
//
//	// 4. Draw the triangle
//	gl.BindVertexArray(cc.VAO)
//	gl.DrawArrays(gl.TRIANGLES, 0, 3)
//	gl.BindVertexArray(0)
//}
