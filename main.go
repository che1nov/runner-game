package main

import (
	"math/rand"
	"strconv"
	"syscall/js"
	"time"
)

// Game представляет состояние игры
type Game struct {
	ctx                 js.Value
	canvasWidth         float64
	canvasHeight        float64
	player              Player
	obstacles           []Obstacle
	score               int
	isGameOver          bool
	playerImage         js.Value
	playerDefeatedImage js.Value
	slotMachineImage    js.Value
}

// Player представляет человека
type Player struct {
	x, y     float64
	width    float64
	height   float64
	velocity float64
	jumping  bool
}

// Obstacle представляет игровой автомат
type Obstacle struct {
	x, y   float64
	width  float64
	height float64
}

func main() {
	// Инициализация canvas
	canvas := js.Global().Get("document").Call("getElementById", "gameCanvas")
	if canvas.IsNull() || canvas.IsUndefined() {
		js.Global().Get("console").Call("error", "Canvas не найден")
		return
	}
	ctx := canvas.Call("getContext", "2d")
	if ctx.IsNull() || ctx.IsUndefined() {
		js.Global().Get("console").Call("error", "Контекст canvas не получен")
		return
	}

	// Получение изображений
	playerImage := js.Global().Get("playerImage")
	playerDefeatedImage := js.Global().Get("playerDefeatedImage")
	slotMachineImage := js.Global().Get("slotMachineImage")
	if playerImage.IsNull() || playerImage.IsUndefined() ||
		playerDefeatedImage.IsNull() || playerDefeatedImage.IsUndefined() ||
		slotMachineImage.IsNull() || slotMachineImage.IsUndefined() {
		js.Global().Get("console").Call("error", "Одно или несколько изображений не загружены")
		return
	}

	// Логирование успешной инициализации
	js.Global().Get("console").Call("log", "Игра инициализирована, изображения загружены")

	// Настройка игры
	game := &Game{
		ctx:          ctx,
		canvasWidth:  800,
		canvasHeight: 400,
		player: Player{
			x:        50,
			y:        325,
			width:    60,
			height:   75,
			velocity: 0,
			jumping:  false,
		},
		obstacles:           []Obstacle{},
		score:               0,
		isGameOver:          false,
		playerImage:         playerImage,
		playerDefeatedImage: playerDefeatedImage,
		slotMachineImage:    slotMachineImage,
	}

	// Обработка нажатия пробела
	js.Global().Get("document").Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		if event.Get("code").String() == "Space" && !game.isGameOver {
			if !game.player.jumping {
				game.player.velocity = -12
				game.player.jumping = true
				js.Global().Get("console").Call("log", "Прыжок")
			}
		}
		return nil
	}))

	// Запуск игрового цикла
	go game.gameLoop()

	// Держим программу активной
	select {}
}

// gameLoop управляет обновлением и отрисовкой
func (g *Game) gameLoop() {
	ticker := time.NewTicker(time.Second / 60) // 60 FPS
	rand.Seed(time.Now().UnixNano())

	for range ticker.C {
		if g.isGameOver {
			g.drawGameOver() // Продолжаем отрисовывать экран конца игры
			continue
		}

		g.update()
		g.draw()
	}
}

// update обновляет состояние игры
func (g *Game) update() {
	// Обновление игрока
	g.player.y += g.player.velocity
	g.player.velocity += 0.5 // Гравитация

	// Проверка земли
	if g.player.y > 325 {
		g.player.y = 325
		g.player.velocity = 0
		g.player.jumping = false
	}

	// Генерация препятствий
	if rand.Float64() < 0.005 {
		g.obstacles = append(g.obstacles, Obstacle{
			x:      g.canvasWidth,
			y:      310,
			width:  45,
			height: 90,
		})
	}

	// Обновление препятствий
	for i := 0; i < len(g.obstacles); i++ {
		g.obstacles[i].x -= 5 // Скорость
	}

	// Удаление препятствий за экраном
	if len(g.obstacles) > 0 && g.obstacles[0].x < -g.obstacles[0].width {
		g.obstacles = g.obstacles[1:]
	}

	// Проверка коллизий
	for _, obs := range g.obstacles {
		if g.checkCollision(g.player, obs) {
			g.isGameOver = true
			js.Global().Get("console").Call("log", "Столкновение, вызываем drawGameOver")
			g.drawGameOver()
			return
		}
	}

	// Обновление счета
	g.score++
}

// draw отрисовывает игру
func (g *Game) draw() {
	// Очистка canvas
	g.ctx.Call("clearRect", 0, 0, g.canvasWidth, g.canvasHeight)

	// Отрисовка надписи "Беги, Артём, беги"
	g.ctx.Set("fillStyle", "darkblue")
	g.ctx.Set("font", "30px Arial")
	g.ctx.Set("textAlign", "center")
	g.ctx.Call("fillText", "Беги, Артём, беги", g.canvasWidth/2, 50)

	// Отрисовка игрока (человек)
	g.ctx.Call("drawImage", g.playerImage, g.player.x, g.player.y, g.player.width, g.player.height)

	// Отрисовка препятствий (игровые автоматы)
	for _, obs := range g.obstacles {
		g.ctx.Call("drawImage", g.slotMachineImage, obs.x, obs.y, obs.width, obs.height)
	}

	// Отрисовка счета
	g.ctx.Set("fillStyle", "black")
	g.ctx.Set("font", "20px Arial")
	g.ctx.Set("textAlign", "left")
	g.ctx.Call("fillText", "Счет: "+strconv.Itoa(g.score), 10, 30)
}

// drawGameOver отображает экран конца игры
func (g *Game) drawGameOver() {
	// Очистка canvas
	g.ctx.Call("clearRect", 0, 0, g.canvasWidth, g.canvasHeight)
	js.Global().Get("console").Call("log", "Выполняется drawGameOver")

	// Отрисовка игрока с изображением поражения (горизонтальное)
	if !g.playerDefeatedImage.IsNull() && !g.playerDefeatedImage.IsUndefined() {
		g.ctx.Call("drawImage", g.playerDefeatedImage, g.player.x, 340, 75, 60)
		js.Global().Get("console").Call("log", "Отрисован player_defeated.png на x="+strconv.FormatFloat(g.player.x, 'f', 2, 64)+", y=340")
	} else {
		// Запасной вариант: красный прямоугольник и текст ошибки
		g.ctx.Set("fillStyle", "red")
		g.ctx.Call("fillRect", g.player.x, 340, 75, 60)
		g.ctx.Set("fillStyle", "white")
		g.ctx.Set("font", "12px Arial")
		g.ctx.Set("textAlign", "center")
		g.ctx.Call("fillText", "No defeated img", g.player.x+37.5, 370)
		js.Global().Get("console").Call("error", "player_defeated.png не загружен, отрисован прямоугольник")
	}

	// Отрисовка препятствий
	for _, obs := range g.obstacles {
		g.ctx.Call("drawImage", g.slotMachineImage, obs.x, obs.y, obs.width, obs.height)
	}

	// Отрисовка текста конца игры
	g.ctx.Set("fillStyle", "red")
	g.ctx.Set("font", "40px Arial")
	g.ctx.Set("textAlign", "center")
	g.ctx.Call("fillText", "Игра окончена!", 250, 200)
	g.ctx.Set("font", "20px Arial")
	g.ctx.Call("fillText", "Счет: "+strconv.Itoa(g.score), 350, 250)
}

// checkCollision проверяет столкновение
func (g *Game) checkCollision(p Player, o Obstacle) bool {
	return p.x < o.x+o.width &&
		p.x+p.width > o.x &&
		p.y < o.y+o.height &&
		p.y+p.height > o.y
}
