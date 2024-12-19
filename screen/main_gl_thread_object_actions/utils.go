package main_gl_thread_object_actions

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.5-core/gl"
)

func compileShaderProgram(vertexSource, fragmentSource string) uint32 {
	// Compile vertex shader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	csource, free := gl.Strs(vertexSource) // Unpack both pointer and cleanup function
	gl.ShaderSource(vertexShader, 1, csource, nil)
	defer free() // Defer cleanup to release the C string memory
	gl.CompileShader(vertexShader)

	// Check for vertex shader compile errors
	var status int32
	gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(vertexShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(vertexShader, logLength, nil, gl.Str(log))
		fmt.Printf("Vertex Shader Compile Error: %s\n", log)
	}

	// Compile fragment shader
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	csource, free = gl.Strs(fragmentSource) // Unpack again for fragment shader
	gl.ShaderSource(fragmentShader, 1, csource, nil)
	defer free() // Defer cleanup to release the C string memory
	gl.CompileShader(fragmentShader)

	// Check for fragment shader compile errors
	gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
		fmt.Printf("Fragment Shader Compile Error: %s\n", log)
	}

	// Link the shader program
	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)

	// Check for linking errors
	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(shaderProgram, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(shaderProgram, logLength, nil, gl.Str(log))
		fmt.Printf("Shader Link Error: %s\n", log)
	}

	// Clean up the compiled shaders after linking
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return shaderProgram
}
