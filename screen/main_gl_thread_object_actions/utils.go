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

// checkGLError decodes OpenGL error codes into human-readable form and panics if an error occurs
func CheckGLError(message string) {
	err := gl.GetError()
	if err != gl.NO_ERROR {
		var errorMessage string
		switch err {
		case gl.INVALID_ENUM:
			errorMessage = "GL_INVALID_ENUM: An unacceptable value is specified for an enumerated argument."
		case gl.INVALID_VALUE:
			errorMessage = "GL_INVALID_VALUE: A numeric argument is out of range."
		case gl.INVALID_OPERATION:
			errorMessage = "GL_INVALID_OPERATION: The specified operation is not allowed in the current state."
		case gl.INVALID_FRAMEBUFFER_OPERATION:
			errorMessage = "GL_INVALID_FRAMEBUFFER_OPERATION: The framebuffer object is not complete."
		case gl.OUT_OF_MEMORY:
			errorMessage = "GL_OUT_OF_MEMORY: There is not enough memory left to execute the command."
		case gl.STACK_UNDERFLOW:
			errorMessage = "GL_STACK_UNDERFLOW: An attempt has been made to perform an operation that would cause an internal stack to underflow."
		case gl.STACK_OVERFLOW:
			errorMessage = "GL_STACK_OVERFLOW: An attempt has been made to perform an operation that would cause an internal stack to overflow."
		default:
			errorMessage = fmt.Sprintf("Unknown OpenGL error code: 0x%X", err)
		}
		panic(fmt.Sprintf("OpenGL Error [%s]: %s (0x%X)", message, errorMessage, err))
	}
}
