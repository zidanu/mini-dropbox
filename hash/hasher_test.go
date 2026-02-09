package hash

import "testing"

func TestComputeFileHash(t *testing.T) {
	tests := []struct {
		name          string
		inputFilepath string
		expectedHash  string
	}{
		{
			name:          "correct result",
			inputFilepath: "testfile1",
			expectedHash:  "625fefc509d337c88b96e7cb954c8f0f529524c46e7c0f13735e1606ed27b678",
		},
	}

	for i, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			hash, err := ComputeFileHash(testCase.inputFilepath)
			if err != nil {
				t.Errorf("Test %v - '%s' FAIL: error %v", i, testCase.name, err)
				return
			}
			if hash != testCase.expectedHash {
				t.Errorf("Test %v - '%s' FAIL: expected hash: %v, actual: %v", i, testCase.name, testCase.expectedHash, hash)
			}
		})
	}
}
