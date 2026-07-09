package auth

import (
	"testing"
	"time"
)

func TestLoginRateLimiter_AllowsUpToLimit(t *testing.T) {
	l := NewLoginRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !l.allow("1.2.3.4") {
			t.Fatalf("попытка %d должна быть разрешена", i+1)
		}
	}
	if l.allow("1.2.3.4") {
		t.Fatal("4-я попытка должна быть отклонена (лимит 3)")
	}
}

func TestLoginRateLimiter_IsolatesByIP(t *testing.T) {
	l := NewLoginRateLimiter(1, time.Minute)

	if !l.allow("1.1.1.1") {
		t.Fatal("первая попытка первого IP должна быть разрешена")
	}
	if !l.allow("2.2.2.2") {
		t.Fatal("лимит должен считаться отдельно для каждого IP")
	}
	if l.allow("1.1.1.1") {
		t.Fatal("второй запрос того же IP должен быть отклонён")
	}
}

func TestLoginRateLimiter_ResetsAfterWindow(t *testing.T) {
	l := NewLoginRateLimiter(1, 20*time.Millisecond)

	if !l.allow("9.9.9.9") {
		t.Fatal("первая попытка должна пройти")
	}
	if l.allow("9.9.9.9") {
		t.Fatal("вторая попытка в том же окне должна быть отклонена")
	}
	time.Sleep(30 * time.Millisecond)
	if !l.allow("9.9.9.9") {
		t.Fatal("после истечения окна попытка снова должна проходить")
	}
}
