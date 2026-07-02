# 01 — Git: настройка и первый коммит

## Установка и настройка
```bash
sudo pacman -S git
git config --global user.name "Vladislav Gnilitskii"
git config --global user.email "твой@email.com"
git config --global init.defaultBranch main
git config --global color.ui auto
```

## Структура проекта
```
asu-soit/
├── README.md
├── .gitignore
├── .env.example
├── docs/          ← Obsidian vault
├── backend/
├── frontend/
├── db/migrations/
└── k8s/
```
```bash
mkdir -p ~/projects/asu-soit
cd ~/projects/asu-soit
mkdir -p docs/templates backend frontend db/migrations k8s
git init
```

## .gitignore
```gitignore
backend/bin/
*.exe *.test *.out
frontend/node_modules/
frontend/dist/
.env
.env.*
!.env.example
docs/.obsidian/workspace.json
docs/.obsidian/cache
docs/.obsidian/plugins/
*.log
logs/
*~
*.swp
```

## Первый коммит
```bash
git status
git add .
git status        # убедиться что всё зелёное
git commit -m "init: структура проекта, .gitignore, README, первая заметка"
git log --oneline # проверить историю
```

## SSH-ключ и подключение к GitHub
```bash
ssh-keygen -t ed25519 -C "твой@email.com"
cat ~/.ssh/id_ed25519.pub   # скопировать и вставить на GitHub
# GitHub → Settings → SSH and GPG keys → New SSH key
ssh -T git@github.com       # проверить: "Hi ...! You've successfully authenticated"
```

### Если remote был добавлен как HTTPS — заменить на SSH
```bash
git remote -v   # проверить текущий remote
git remote set-url origin git@github.com:твой-ник/asu-soit.git
git push -u origin main
```

## Ежедневный рабочий цикл
```bash
git pull origin main                  # начало сессии
# ... работа ...
git status
git add .
git commit -m "prefix: описание"
git push
```

## Префиксы коммит-сообщений

| Префикс | Когда использовать |
|---|---|
| `init:` | первый коммит, инициализация |
| `feat:` | новая функциональность |
| `fix:` | исправление бага |
| `docs:` | заметки, документация |
| `db:` | миграции, схема БД |
| `refactor:` | переработка кода без изменения поведения |
| `chore:` | обновление .gitignore, зависимостей и т.п. |

## Ветки
```bash
git checkout -b feature/название   # создать и переключиться
git checkout main                   # вернуться в main
git merge feature/название          # влить ветку в main
git branch -d feature/название      # удалить после слияния
```

## Связанные заметки
- [[00-overview]]
- [[02-database-design]]