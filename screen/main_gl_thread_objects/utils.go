/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package main_gl_thread_objects

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.5-core/gl"
)

// compileShaderProgram takes pointers to C-style uint8 strings for vertex and fragment shader sources.
// It compiles the shaders, links them into a shader program, and returns the program ID.
func compileShaderProgram(vertexSource, fragmentSource *uint8) uint32 {
	// Compile vertex shader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	CheckGLError("After CreateShader (vertex)")
	gl.ShaderSource(vertexShader, 1, &vertexSource, nil)
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
	CheckGLError("After CreateShader (fragment)")
	gl.ShaderSource(fragmentShader, 1, &fragmentSource, nil)
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
	CheckGLError("After CreateProgram")
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

	if !gl.IsProgram(shaderProgram) {
		fmt.Printf("Invalid shader program: %d\n", shaderProgram)
		panic("Shader program validation failed")
	}

	return shaderProgram
}

func setShaderProgram(shaderProgram uint32) {
	if !gl.IsProgram(shaderProgram) {
		fmt.Printf("[Render] Shader program %d is not valid.\n", shaderProgram)
		panic("[Render] Invalid shader program")
	}
	gl.UseProgram(shaderProgram)
	CheckGLError("After UseProgram")
	// Check if the active program matches
	var activeProgram int32
	gl.GetIntegerv(gl.CURRENT_PROGRAM, &activeProgram)
	if uint32(activeProgram) != shaderProgram {
		fmt.Printf("[Render] Shader program mismatch! Visible: %d, "+
			"Expected: %d\n", activeProgram, shaderProgram)
		panic("[Render] Shader program is not active as expected")
	}
	if shaderProgram == 0 {
		fmt.Println("[Render] Shader program handle is 0. " +
			"Possible compilation/linking failure.")
		panic("[Render] Shader program handle is 0")
	}
}

// CheckGLError checkGLError decodes OpenGL error codes into human-readable form and panics if an error occurs
func CheckGLError(message string) {
	err := gl.GetError()
	if err != gl.NO_ERROR {
		var errorMessage string
		switch err {
		case gl.INVALID_ENUM:
			errorMessage = "GL_INVALID_ENUM: An unacceptable value is" +
				" specified  for an enumerated argument."
		case gl.INVALID_VALUE:
			errorMessage = "GL_INVALID_VALUE: A numeric argument is out of" +
				"  range."
		case gl.INVALID_OPERATION:
			errorMessage = "GL_INVALID_OPERATION: The specified operation" +
				" is not allowed in the current state."
		case gl.INVALID_FRAMEBUFFER_OPERATION:
			errorMessage = "GL_INVALID_FRAMEBUFFER_OPERATION: The" +
				" framebuffer object is not complete."
		case gl.OUT_OF_MEMORY:
			errorMessage = "GL_OUT_OF_MEMORY: There is not enough memory" +
				" left to execute the command."
		case gl.STACK_UNDERFLOW:
			errorMessage = "GL_STACK_UNDERFLOW: An attempt has been made to" +
				" perform an operation that would cause an internal stack to" +
				" underflow."
		case gl.STACK_OVERFLOW:
			errorMessage = "GL_STACK_OVERFLOW: An attempt has been made to" +
				" perform an operation that would cause an internal stack to overflow."
		default:
			errorMessage = fmt.Sprintf("Unknown OpenGL error code: 0x%X", err)
		}
		panic(fmt.Sprintf("OpenGL Error [%s]: %s (0x%X)", message,
			errorMessage, err))
	}
}
