package agent

import (
	"testing"
	"time"
)

func TestSignPayload(t *testing.T) {
	payload := &JobPayload{
		JobID:        "test-job-1",
		Prompt:       "test prompt",
		Command:      "echo test",
		TTL:          time.Now().Add(5 * time.Minute).Unix(),
		Timestamp:    time.Now().Unix(),
		ControllerID: "test-controller",
	}
	
	secret := "test-secret-key"
	
	signature, err := SignPayload(payload, secret)
	if err != nil {
		t.Fatalf("SignPayload() error = %v", err)
	}
	
	if signature == "" {
		t.Error("SignPayload() returned empty signature")
	}
	
	// Verify signature length (HMAC-SHA256 produces 64 hex characters)
	if len(signature) != 64 {
		t.Errorf("SignPayload() signature length = %d, want 64", len(signature))
	}
}

func TestValidateSignature(t *testing.T) {
	payload := &JobPayload{
		JobID:        "test-job-1",
		Prompt:       "test prompt",
		Command:      "echo test",
		TTL:          time.Now().Add(5 * time.Minute).Unix(),
		Timestamp:    time.Now().Unix(),
		ControllerID: "test-controller",
	}
	
	secret := "test-secret-key"
	
	signature, _ := SignPayload(payload, secret)
	
	signedJob := &SignedJob{
		Payload: *payload,
		Signature: JobSignature{
			Signature: signature,
			Algorithm: "HMAC-SHA256",
		},
	}
	
	// Valid signature
	err := ValidateSignature(signedJob, secret)
	if err != nil {
		t.Errorf("ValidateSignature() error = %v, want nil", err)
	}
	
	// Invalid signature
	signedJob.Signature.Signature = "invalid-signature"
	err = ValidateSignature(signedJob, secret)
	if err != ErrInvalidSignature {
		t.Errorf("ValidateSignature() error = %v, want ErrInvalidSignature", err)
	}
	
	// Wrong secret
	signedJob.Signature.Signature = signature
	err = ValidateSignature(signedJob, "wrong-secret")
	if err != ErrInvalidSignature {
		t.Errorf("ValidateSignature() with wrong secret error = %v, want ErrInvalidSignature", err)
	}
}

func TestValidateTTL(t *testing.T) {
	tests := []struct {
		name      string
		payload   *JobPayload
		wantError error
	}{
		{
			name: "Valid TTL",
			payload: &JobPayload{
				TTL:       time.Now().Add(5 * time.Minute).Unix(),
				Timestamp: time.Now().Unix(),
			},
			wantError: nil,
		},
		{
			name: "Expired TTL",
			payload: &JobPayload{
				TTL:       time.Now().Add(-1 * time.Minute).Unix(),
				Timestamp: time.Now().Unix(),
			},
			wantError: ErrJobExpired,
		},
		{
			name: "Timestamp too old (replay attack)",
			payload: &JobPayload{
				TTL:       time.Now().Add(5 * time.Minute).Unix(),
				Timestamp: time.Now().Add(-10 * time.Minute).Unix(),
			},
			wantError: ErrJobTooOld,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTTL(tt.payload)
			if err != tt.wantError {
				t.Errorf("ValidateTTL() error = %v, want %v", err, tt.wantError)
			}
		})
	}
}

func TestJobPayloadSerialization(t *testing.T) {
	payload := &JobPayload{
		JobID:        "test-job-1",
		Prompt:       "test prompt",
		Command:      "echo test",
		CandidateMetadata: map[string]interface{}{
			"confidence": 95,
			"risk_level": "safe",
		},
		PluginMetadata: map[string]interface{}{
			"plugin_name": "git",
		},
		RequiredScopes: []string{"git:read", "git:write"},
		TTL:            time.Now().Add(5 * time.Minute).Unix(),
		Timestamp:      time.Now().Unix(),
		ControllerID:   "test-controller",
	}
	
	// Sign payload
	secret := "test-secret"
	signature, err := SignPayload(payload, secret)
	if err != nil {
		t.Fatalf("SignPayload() error = %v", err)
	}
	
	signedJob := &SignedJob{
		Payload: *payload,
		Signature: JobSignature{
			Signature: signature,
			Algorithm: "HMAC-SHA256",
		},
	}
	
	// Validate signature
	if err := ValidateSignature(signedJob, secret); err != nil {
		t.Errorf("ValidateSignature() error = %v", err)
	}
	
	// Validate TTL
	if err := ValidateTTL(&signedJob.Payload); err != nil {
		t.Errorf("ValidateTTL() error = %v", err)
	}
}
