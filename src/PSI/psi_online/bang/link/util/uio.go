package util

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func WriteUint64ToFile(filename string, value uint64) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%d", value)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func ReadUint64FromFile(filename string) (uint64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return 0, fmt.Errorf("error scanning file: %w", err)
		}
		return 0, nil
	}

	value, err := strconv.ParseUint(scanner.Text(), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing value: %w", err)
	}

	return value, nil
}
