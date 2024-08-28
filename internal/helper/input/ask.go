// Package input /*
/*
Copyright © 2024 Benek <benek2048@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package input

import (
	"bufio"
	"os"
	"strings"
)

// AskForYesOrNot is a function that prompts the user for a yes or no response in the console.
// This function is useful for getting user confirmation in console applications.
//
// Parameters:
//   - answerYes: The expected answer for a yes response. This parameter is case-insensitive.
//   - defaultAnswer: The default answer that will be used if the user provides no input. This parameter is case-insensitive.
//
// Returns:
//   - answer: A boolean indicating whether the user's input matches the yes answer.
//
// Example:
//
//	AskForYesOrNot("yes", "no") will return true if the user inputs "yes", and false otherwise.
func AskForYesOrNot(answerYes string, defaultAnswer string) (answer bool) {
	input := getInput()
	if input == "" {
		input = strings.ToLower(defaultAnswer)
	}
	answer = input == strings.ToLower(answerYes)
	return
}

// AskForYesOrNotOrForAll is a function that prompts the user for a yes, no, or all response in the console.
// This function is useful for getting user confirmation in console applications, with an additional option for "all".
//
// Parameters:
//   - answerYes: The expected answer for a yes response. This parameter is case-insensitive.
//   - answerForAll: The expected answer for an "all" response. This parameter is case-insensitive.
//   - defaultAnswer: The default answer that will be used if the user provides no input. This parameter is case-insensitive.
//
// Returns:
//   - answer: A boolean indicating whether the user's input matches the default answer or the "all" answer.
//   - forAll: A boolean indicating whether the user's input matches the "all" answer.
//
// Example:
//
//	AskForYesOrNotOrForAll("yes", "all", "no") will return true if the user inputs "yes" or "all", and false otherwise.
//	It will also return true for the second return value if the user inputs "all", and false otherwise.
func AskForYesOrNotOrForAll(answerYes string, answerForAll string, defaultAnswer string) (answer bool, forAll bool) {
	input := getInput()
	if input == "" {
		input = strings.ToLower(defaultAnswer)
	}
	forAll = input == strings.ToLower(answerForAll)
	answer = forAll || input == strings.ToLower(answerYes)
	return
}

// getInput reads a line of input from the user in the console and returns it as a lowercase string with whitespace trimmed.
func getInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(input))
}
