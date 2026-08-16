package util

import "testing"

func TestStack(t *testing.T) {
	t.Run("lifo ordering", func(t *testing.T) {
		stack := Stack[int]{}
		for _, v := range []int{1, 2, 3} {
			stack.Push(v)
		}

		if stack.Len() != 3 {
			t.Fatalf("Len = %d, want 3", stack.Len())
		}
		if got := stack.Pop(); got != 3 {
			t.Errorf("Pop = %d, want 3", got)
		}
		if got := stack.Pop(); got != 2 {
			t.Errorf("Pop = %d, want 2", got)
		}
		if got := stack.Pop(); got != 1 {
			t.Errorf("Pop = %d, want 1", got)
		}
		if stack.Len() != 0 {
			t.Errorf("Len = %d, want 0", stack.Len())
		}
	})

	t.Run("peek does not remove", func(t *testing.T) {
		stack := Stack[int]{}
		stack.Push(7)
		stack.Push(8)

		if got := stack.Peek(); got != 8 {
			t.Errorf("Peek = %d, want 8", got)
		}
		if stack.Len() != 2 {
			t.Errorf("Len = %d, want 2 after peek", stack.Len())
		}
	})
}
