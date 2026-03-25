package reyna

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/reyna-bot/reyna-backend/internal/models"
)

// Reyna is the personality engine for the bot
type Reyna struct{}

func New() *Reyna { return &Reyna{} }

func pick(opts []string) string {
	return opts[rand.Intn(len(opts))]
}

func isLateNight() bool {
	h := time.Now().Hour()
	return h >= 23 || h < 5
}

func isExamSeason() bool {
	m := time.Now().Month()
	return m == time.November || m == time.December || m == time.April || m == time.May
}

// AddResponse generates Reyna's response after storing a file
func (r *Reyna) AddResponse(fileName string, version int, totalFiles int) string {
	base := []string{
		fmt.Sprintf("Saved `%s` → your Drive. That's %d files total. Tera Drive toh library ban raha hai 📚", fileName, totalFiles),
		fmt.Sprintf("Done! `%s` stored ✅ Ab ye file safe hai — unlike tera exam preparation 📖", fileName),
		fmt.Sprintf("`%s` saved. Isse zyada organized toh tera room bhi nahi hoga 🗂️", fileName),
		fmt.Sprintf("Stored `%s` → Drive. Aur kitne bhejega? Free mein kaam karti hoon but unlimited nahi hoon 😤", fileName),
	}

	if version > 1 {
		base = append(base,
			fmt.Sprintf("`%s` v%d saved. Itne versions toh Hollywood movies ke nahi hote 🎬", fileName, version),
			fmt.Sprintf("Updated to v%d! Pehle wala bhi safe hai. Main version control karti hoon, delete nahi ✅", version),
		)
	}

	resp := pick(base)

	if isLateNight() {
		resp += "\n\n_(Btw, raat ke " + fmt.Sprintf("%d", time.Now().Hour()) + " baj rahe hain. So ja bhai.)_ 🌙"
	}

	return resp
}

// FindResponse generates Reyna's response for search results
func (r *Reyna) FindResponse(query string, files []models.File) string {
	if len(files) == 0 {
		return pick([]string{
			fmt.Sprintf("🔍 \"%s\" se kuch nahi mila. Ya toh aisa file exist nahi karta ya tum log share karna bhool gaye. Classic. 🤷‍♀️", query),
			fmt.Sprintf("Zero results for \"%s\". Group mein notes daalte toh milte na? Memes toh bohot hain. 😂", query),
			fmt.Sprintf("Nothing found for \"%s\". Koi notes share karo pehle, phir search karo. Logic. 🧠", query),
		})
	}

	result := fmt.Sprintf("🔍 Found %d result(s) for \"%s\":\n", len(files), query)
	for i, f := range files {
		if i >= 5 {
			result += fmt.Sprintf("...and %d more\n", len(files)-5)
			break
		}
		result += fmt.Sprintf("📄 %s (v%d, by %s)\n", f.FileName, f.Version, f.SharedByName)
	}

	if isLateNight() {
		result += pick([]string{
			"\nRaat ke 2 baje search? Exam kal hai na? Classic engineering move 🎯",
			"\nBhai so ja. Notes subah bhi milenge. Ya nahi milenge because tu padh ke bhi kuch nahi karega 😴",
		})
	} else if isExamSeason() {
		result += "\n_Exam season vibes detected. Sab suddenly scholar ban rahe hain_ 📚"
	}

	return result
}

// LogResponse generates Reyna's response for file history
func (r *Reyna) LogResponse(files []models.File, total int) string {
	if total == 0 {
		return pick([]string{
			"📋 Empty log. Koi kuch share nahi karta kya is group mein? Sad. 😢",
			"📋 Zero files. Ye group hai ya ghost town? Share something first. 👻",
		})
	}

	result := fmt.Sprintf("📋 Last %d entries (of %d total):\n", min(len(files), 10), total)
	for i, f := range files {
		if i >= 10 {
			break
		}
		ago := time.Since(f.CreatedAt)
		agoStr := formatDuration(ago)
		result += fmt.Sprintf("• %s — %s ago (by %s)\n", f.FileName, agoStr, f.SharedByName)
	}

	result += pick([]string{
		fmt.Sprintf("\n%d files total. Group mein notes zyada hain ya memes? 🤔", total),
		fmt.Sprintf("\n%d total. Tu toh version control pro ban gaya re 🚀", total),
	})

	return result
}

// StatusResponse generates Reyna's status update
func (r *Reyna) StatusResponse(newFiles []models.File, total int) string {
	if len(newFiles) == 0 {
		return pick([]string{
			"📊 Nothing new. Group mein sab so rahe hain kya? Koi notes daalo yaar. 😴",
			"📊 Status: All quiet. No new files. Everyone's either chill or lazy. Probably lazy. 🦥",
		})
	}

	subjects := make(map[string]int)
	for _, f := range newFiles {
		sub := f.Subject
		if sub == "" {
			sub = "misc"
		}
		subjects[sub]++
	}

	result := fmt.Sprintf("📊 %d new file(s) since your last check:\n", len(newFiles))
	for sub, cnt := range subjects {
		result += fmt.Sprintf("• %s: %d file(s)\n", sub, cnt)
	}

	if isExamSeason() {
		result += "\n_Exam season incoming. Suddenly everyone remembers this group exists_ 📚😂"
	}

	result += pick([]string{
		fmt.Sprintf("\nTotal repo: %d files. Keep syncing. 🔄", total),
		fmt.Sprintf("\n%d total files stored. Tera group sabse organized hai campus mein 💪", total),
	})

	return result
}

// HelpResponse generates Reyna's help text
func (r *Reyna) HelpResponse() string {
	return pick([]string{
		`🤖 Reyna hoon — tera group ka sabse useful member.

/reyna add .        → last file stage kar
/reyna add File.pdf → specific file stage kar
/reyna staged       → staged files dekh
/reyna commit       → staged files → Drive push
/reyna commit File  → specific file commit
/reyna rm File      → staged file remove
/reyna rm .         → sab staged remove
/reyna find "xyz"   → search kar
/reyna log          → history dekh
/reyna status       → kya naya hai

Basically, main wo kaam karti hoon jo tum log nahi karte — organize 📁`,

		`Help chahiye? Simple:

/reyna add → file stage
/reyna commit → Drive mein push
/reyna rm → staged se remove
/reyna find → dhundho
/reyna log → history
/reyna status → updates
/reyna staged → staging area dekh

Itna toh ChatGPT bhi bata deta. But ChatGPT mein attitude nahi hai 💅`,

		`Main karu toh kya karu — saves your files like git.

add → stage. commit → push to Drive. rm → unstage.
find, log, status, staged — sab hai.

Ab padh le bhai. 📖`,
	})
}

// GenericResponse for non-command messages
func (r *Reyna) GenericResponse() string {
	return pick([]string{
		"Bhai main bot hoon, tutor nahi. Commands de, gyaan mat. Try /reyna help 🙃",
		"Ye WhatsApp group hai, therapy session nahi. Use /reyna help 😅",
		"Interesting message. Filing under 'things Reyna doesn't care about.' Try /reyna help 🤷‍♀️",
		"Sorry, meri range mein nahi aata ye. Main files save karti hoon, emotional support nahi deti. /reyna help 💁‍♀️",
		"Bhai tera message padha. Kuch samajh nahi aaya. /reyna help try kar. 🫠",
	})
}

// ── Commit Responses ──

func (r *Reyna) CommitFileResponse(fileName string) string {
	return pick([]string{
		fmt.Sprintf("✅ `%s` committed! Ab ye permanent hai — Google Drive mein ja raha hai. No going back 🚀", fileName),
		fmt.Sprintf("Committed `%s` ✅ Ab ye file pakka saved hai. Tera semester bach gaya 🎓", fileName),
	})
}

func (r *Reyna) CommitAllResponse(count int) string {
	return pick([]string{
		fmt.Sprintf("✅ %d file(s) committed! Sab kuch Drive mein push ho gaya. `git push origin main` energy 🚀", count),
		fmt.Sprintf("Boom! %d files committed. Ab ye permanent hain. Exam ke time thank karega mujhe 😎", count),
		fmt.Sprintf("%d files committed ✅ Staged area ab khali hai. Clean slate, just like tera attendance record 📋", count),
	})
}

func (r *Reyna) CommitEmptyResponse() string {
	return pick([]string{
		"⚠️ Kuch commit karne ko hai hi nahi. Pehle `/reyna add` se files stage kar, phir commit kar.",
		"Staged area empty hai bhai. Add files first → then commit. Basic git workflow 🤷‍♀️",
	})
}

func (r *Reyna) CommitError(fileName string) string {
	return fmt.Sprintf("❌ `%s` commit nahi ho paaya. Check if it's staged. Use `/reyna staged` to see.", fileName)
}

// ── Rm Responses ──

func (r *Reyna) RmFileResponse(fileName string) string {
	return pick([]string{
		fmt.Sprintf("🗑️ `%s` removed from staging. Poof — gone. Like tera last relationship 💨", fileName),
		fmt.Sprintf("Deleted `%s` from staged files ✅ Ab wo file exist nahi karti. Denial nahi, reality 🗑️", fileName),
	})
}

func (r *Reyna) RmAllResponse(count int) string {
	return pick([]string{
		fmt.Sprintf("🗑️ %d staged file(s) removed. Clean slate. Ab fresh start kar, like new semester 📚", count),
		fmt.Sprintf("Sabse clean kiya — %d files gone from staging. `git reset --hard HEAD` vibes 🧹", count),
	})
}

func (r *Reyna) RmEmptyResponse() string {
	return pick([]string{
		"Staging area already empty hai. Kya remove karu? Hawa? 💨",
		"Nothing to remove. Pehle kuch add kar, phir remove karna. Logic. 🧠",
	})
}

func (r *Reyna) RmNotFoundResponse(fileName string) string {
	return fmt.Sprintf("❌ `%s` staged mein nahi mili. Ya toh already committed hai ya exist hi nahi karti. `/reyna staged` check kar.", fileName)
}

// ── Staged Response ──

func (r *Reyna) StagedResponse(files []models.File) string {
	if len(files) == 0 {
		return pick([]string{
			"📋 Staging area empty. Koi file staged nahi hai. Use `/reyna add` to stage files.",
			"Nothing staged. Sab ya toh committed hain ya kisi ne add hi nahi kiya. Typical. 🤷‍♀️",
		})
	}

	result := fmt.Sprintf("📋 %d file(s) staged (not yet committed):\n", len(files))
	for i, f := range files {
		if i >= 10 {
			result += fmt.Sprintf("...and %d more\n", len(files)-10)
			break
		}
		result += fmt.Sprintf("  📄 %s (%s) — by %s\n", f.FileName, f.Subject, f.SharedByName)
	}
	result += "\nUse `/reyna commit` to push to Drive, or `/reyna rm <file>` to remove."
	return result
}

// DuplicateWarning when file already exists
func (r *Reyna) DuplicateWarning(fileName string, count int) string {
	return pick([]string{
		fmt.Sprintf("⚠️ `%s` already %d baar aa chuki hai. %dth time share karne se marks nahi badhenge. Stored as new version anyway ✅", fileName, count, count+1),
		fmt.Sprintf("`%s` phir se? Bhai ye file toh veteran hai is group ki. v%d saved. 🔄", fileName, count+1),
	})
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ProcessCommand is the main entry point for WhatsApp commands
func (r *Reyna) ProcessCommand(cmd string) (action string, args string) {
	cmd = strings.TrimSpace(cmd)
	if !strings.HasPrefix(strings.ToLower(cmd), "/reyna") {
		return "unknown", cmd
	}

	parts := strings.Fields(cmd)
	if len(parts) < 2 {
		return "help", ""
	}

	action = strings.ToLower(parts[1])
	if len(parts) > 2 {
		args = strings.Join(parts[2:], " ")
		args = strings.Trim(args, "\"'")
	}

	switch action {
	case "add", "find", "log", "status", "help", "commit", "rm", "staged":
		return action, args
	default:
		return "unknown", cmd
	}
}
