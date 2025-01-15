/*
 * // This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
 * // If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
 * // 2024
 */

package screen

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.5-core/gl"
)

var DEBUG = false

// compileShaderProgram takes pointers to C-style uint8 strings for vertex and fragment shader sources.
// It compiles the shaders, links them into a shader program, and returns the program ID.
func compileShaderProgram(vertexSource, fragmentSource,
	geomSource *uint8) uint32 {
	// Compile vertex shader
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	CheckGLError("After CreateShader (vertex)")
	gl.ShaderSource(vertexShader, 1, &vertexSource, nil)
	gl.CompileShader(vertexShader)
	var status int32
	if DEBUG {
		// Check for vertex shader compile errors
		gl.GetShaderiv(vertexShader, gl.COMPILE_STATUS, &status)
		if status == gl.FALSE {
			var logLength int32
			gl.GetShaderiv(vertexShader, gl.INFO_LOG_LENGTH, &logLength)
			log := strings.Repeat("\x00", int(logLength+1))
			gl.GetShaderInfoLog(vertexShader, logLength, nil, gl.Str(log))
			fmt.Printf("Vertex Shader Compile Error: %s\n", log)
		}
	}

	// Compile fragment shader
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	CheckGLError("After CreateShader (fragment)")
	gl.ShaderSource(fragmentShader, 1, &fragmentSource, nil)
	gl.CompileShader(fragmentShader)

	if DEBUG {
		// Check for fragment shader compile errors
		gl.GetShaderiv(fragmentShader, gl.COMPILE_STATUS, &status)
		if status == gl.FALSE {
			var logLength int32
			gl.GetShaderiv(fragmentShader, gl.INFO_LOG_LENGTH, &logLength)
			log := strings.Repeat("\x00", int(logLength+1))
			gl.GetShaderInfoLog(fragmentShader, logLength, nil, gl.Str(log))
			fmt.Printf("Fragment Shader Compile Error: %s\n", log)
		}
	}

	var geomShader uint32
	if geomSource != nil {
		// Compile fragment shader
		geomShader = gl.CreateShader(gl.GEOMETRY_SHADER)
		CheckGLError("After CreateShader (GEOM)")
		gl.ShaderSource(geomShader, 1, &geomSource, nil)
		gl.CompileShader(geomShader)

		if DEBUG {
			// Check for fragment shader compile errors
			gl.GetShaderiv(geomShader, gl.COMPILE_STATUS, &status)
			if status == gl.FALSE {
				var logLength int32
				gl.GetShaderiv(geomShader, gl.INFO_LOG_LENGTH, &logLength)
				log := strings.Repeat("\x00", int(logLength+1))
				gl.GetShaderInfoLog(geomShader, logLength, nil, gl.Str(log))
				fmt.Printf("Geom Shader Compile Error: %s\n", log)
			}
		}
	}

	// Link the shader program
	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	if geomSource != nil {
		gl.AttachShader(shaderProgram, geomShader)
	}
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)
	CheckGLError("After LinkProgram")

	// Check for linking errors
	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(shaderProgram, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(shaderProgram, logLength, nil, gl.Str(log))
		fmt.Printf("Shader Link Error: %s\n", log)

	}
	if !gl.IsProgram(shaderProgram) {
		fmt.Printf("Invalid shader program: %d\n", shaderProgram)
		panic("Shader program validation failed")
	}
	gl.ValidateProgram(shaderProgram)
	var validateStatus int32
	gl.GetProgramiv(shaderProgram, gl.VALIDATE_STATUS, &validateStatus)
	if validateStatus == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(shaderProgram, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(shaderProgram, logLength, nil, gl.Str(log))
		panic(fmt.Sprintf("Program validation failed: %s", log))
	}

	// Clean up the compiled shaders after linking
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)
	if geomSource != nil {
		gl.DeleteShader(geomShader)
	}

	return shaderProgram
}

func setShaderProgram(shaderProgram uint32) {
	if DEBUG {
		CheckGLError("Before SetShaderProgram")
		if !gl.IsProgram(shaderProgram) {
			fmt.Printf("[render] Shader program %d is not valid.\n", shaderProgram)
			panic("[render] Invalid shader program")
		}
	}
	gl.UseProgram(shaderProgram)
	CheckGLError("After UseProgram")
	if DEBUG {
		// Check if the active program matches
		var activeProgram int32
		gl.GetIntegerv(gl.CURRENT_PROGRAM, &activeProgram)
		if uint32(activeProgram) != shaderProgram {
			fmt.Printf("[render] Shader program mismatch! Visible: %d, "+
				"Expected: %d\n", activeProgram, shaderProgram)
			panic("[render] Shader program is not active as expected")
		}
		if shaderProgram == 0 {
			fmt.Println("[render] Shader program handle is 0. " +
				"Possible compilation/linking failure.")
			panic("[render] Shader program handle is 0")
		}
	}
}

// CheckGLError checkGLError decodes OpenGL error codes into human-readable form and panics if an error occurs
func CheckGLError(message string) {
	errCode := gl.GetError()
	if errCode != gl.NO_ERROR {
		fmt.Printf("%s: ", message)
		switch errCode {
		case gl.INVALID_ENUM:
			panic("GL_INVALID_ENUM: An unacceptable value is" +
				" specified  for an enumerated argument.")
		case gl.INVALID_VALUE:
			panic("GL_INVALID_VALUE: A numeric argument is out of" +
				"  range.")
		case gl.INVALID_OPERATION:
			panic("GL_INVALID_OPERATION: The specified operation" +
				" is not allowed in the current state.")
		case gl.INVALID_FRAMEBUFFER_OPERATION:
			panic("GL_INVALID_FRAMEBUFFER_OPERATION: The" +
				" framebuffer object is not complete.")
		case gl.OUT_OF_MEMORY:
			panic("GL_OUT_OF_MEMORY: There is not enough memory" +
				" left to execute the command.")
		case gl.STACK_UNDERFLOW:
			panic("GL_STACK_UNDERFLOW: An attempt has been made to" +
				" perform an operation that would cause an internal stack to" +
				" underflow.")
		case gl.STACK_OVERFLOW:
			panic("GL_STACK_OVERFLOW: An attempt has been made to" +
				" perform an operation that would cause an internal stack to" +
				" overflow.")
		default:
			panic("Unknown OpenGL error code")
		}
	}
}
