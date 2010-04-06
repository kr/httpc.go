package httpc

import (
	"io/ioutil"
	"testing"
)

func TestSet(t *testing.T) {
	body := []byte("body")
	s := NewMemoryStore(7)
	s.Set("a", map[string]string{}, body)
	info, content := s.Get("a")
	if info == nil {
		t.Error("expected info, got nil")
	}
	if content == nil {
		t.Fatal("expected content, got nil")
	}
	x, err := ioutil.ReadAll(content)
	if err != nil {
		t.Error("unexpected err", err)
	}
	if string(x) != string(body) {
		t.Errorf("expected body %#v, got %#v", body, string(x))
	}
}

func TestLimit(t *testing.T) {
	body := []byte("long body that doesn't fit")
	s := NewMemoryStore(7)
	s.Set("a", map[string]string{}, body)
	info, content := s.Get("a")
	if info != nil {
		t.Errorf("expected nil info, got %#v", info)
	}
	if content != nil {
		t.Errorf("expected nil content, got %#v", content)
	}
}

func TestReplacement(t *testing.T) {
	body1 := []byte("body1")
	body2 := []byte("body2")
	s := NewMemoryStore(7)
	s.Set("a", map[string]string{}, body1)
	s.Set("b", map[string]string{}, body2)

	// Check that the first entry got kicked out.
	info1, content1 := s.Get("a")
	if info1 != nil {
		t.Errorf("expected nil info1, got %#v", info1)
	}
	if content1 != nil {
		t.Errorf("expected nil content1, got %#v", content1)
	}

	// Check that the second entry got stored.
	info2, content2 := s.Get("b")
	if info2 == nil {
		t.Error("expected info2, got nil")
	}
	if content2 == nil {
		t.Fatal("expected content2, got nil")
	}
	x, err := ioutil.ReadAll(content2)
	if err != nil {
		t.Error("unexpected err", err)
	}
	if string(x) != string(body2) {
		t.Errorf("expected body %#v, got %#v", string(body2), string(x))
	}
}

