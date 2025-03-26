package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/tebeka/selenium"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	if username == "" {
		http.Error(w, "username parameter is required", http.StatusBadRequest)
		return
	}
	if password == "" {
		http.Error(w, "password parameter is required", http.StatusBadRequest)
		return
	}

	service, err := selenium.NewChromeDriverService("./chromedriver", 4444)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	defer service.Stop()

	caps := selenium.Capabilities{"browserName": "chrome"}
	// chromeCaps := chrome.Capabilities{
	// 	Args: []string{
	// 		"--headless",
	// 	},
	// }
	// caps.AddChrome(chromeCaps)

	// caps.AddChrome(chrome.Capabilities{Path: "./chrome-linux64/chrome", Args: []string{
	// 	"--headless",
	// }})

	driver, err := selenium.NewRemote(caps, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = driver.MaximizeWindow("")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	authFlow(w, driver, username, password)

	allCookies, err := driver.GetCookies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	// fileAllCookies, err := os.Create("allCookies.json")
	// if err != nil {
	// http.Error(err)
	// }
	// defer fileAllCookies.Close()
	// encoder := json.NewEncoder(fileAllCookies)
	// encoder.SetIndent("", "  ")
	// err = encoder.Encode(allCookies)
	// if err != nil {
	// http.Error(err)
	// }
	// fmt.Println("Успешно сохранили Cookies в allCookies.json")

	w.Header().Set("Content-Type", "application/json")
	// Кодируем структуру в JSON и отправляем
	if err := json.NewEncoder(w).Encode(allCookies); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func authFlow(w http.ResponseWriter, driver selenium.WebDriver, username, password string) {
	err := driver.Get("https://www.threads.net/login/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// driver.SetPageLoadTimeout(100 * time.Second)

	// pageScreenshot(driver, "screen1")
	time.Sleep(2 * time.Second)
	// pageScreenshot(driver, "screen2")
	time.Sleep(1 * time.Second)

	acceptAllCookies(driver)

	time.Sleep(2 * time.Second)
	// pageScreenshot(driver, "screen3")

	continueWithInstagram(driver)

	time.Sleep(2 * time.Second)
	// pageScreenshot(driver, "screen4")

	fillCredsAndLogin(driver, username, password)

	time.Sleep(10 * time.Second)
	// pageScreenshot(driver, "screen7")

	//get cookies
	// getAllCookies(driver)
}

func cryptoRandom(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return int(n.Int64()) + min
}

func acceptAllCookies(driver selenium.WebDriver) {
	var elemCookieAccept selenium.WebElement
	//find with waiting
	err := driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		foundElem, err := driver.FindElement(selenium.ByXPATH, "//div[@role='button' and .//div[contains(text(), 'Разрешить все cookie')]]")
		if err != nil {
			panic(fmt.Errorf("не удалось найти кнопку 'Разрешить все cookie': %v", err))
		}
		elemCookieAccept = foundElem
		visible, err := foundElem.IsDisplayed()
		return visible, err
	}, 10*time.Second)
	if err != nil {
		panic(fmt.Errorf("не удалось найти элемент: %v", err))
	}

	//scroll to element
	driver.ExecuteScript("arguments[0].scrollIntoView({block: 'center'});", []interface{}{elemCookieAccept})

	//click
	time.Sleep(time.Duration(cryptoRandom(300, 500)) * time.Millisecond)
	_, err = driver.ExecuteScript("arguments[0].click();", []interface{}{elemCookieAccept})
	if err != nil {
		panic(fmt.Errorf("не удалось кликнуть по кнопке 'Разрешить все cookie': %v", err))
	}
	fmt.Println("Успешно нажали на 'Разрешить все cookie'")
}

func continueWithInstagram(driver selenium.WebDriver) {
	//find with waiting
	var elemContinueWithInstagram selenium.WebElement
	err := driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		foundElem, err := wd.FindElement(selenium.ByXPATH, "//a[.//span[contains(text(), 'Продолжить с аккаунтом Instagram')]]")
		if err != nil {
			panic(fmt.Errorf("не удалось найти кнопку 'Продолжить с аккаунтом Instagram': %v", err))
		}
		elemContinueWithInstagram = foundElem
		visible, err := foundElem.IsDisplayed()
		return visible, err
	}, 10*time.Second)
	if err != nil {
		panic(fmt.Errorf("не удалось найти элемент: %v", err))
	}

	//scroll to element
	time.Sleep(time.Duration(cryptoRandom(300, 500)) * time.Millisecond)
	driver.ExecuteScript("arguments[0].scrollIntoView(true);", []interface{}{elemContinueWithInstagram})

	//click
	time.Sleep(time.Duration(cryptoRandom(300, 500)) * time.Millisecond)
	_, err = driver.ExecuteScript("arguments[0].click();", []interface{}{elemContinueWithInstagram})
	if err != nil {
		panic(fmt.Errorf("не удалось кликнуть по кнопке 'Продолжить с аккаунтом Instagram': %v", err))
	}
	fmt.Println("Успешно нажали на 'Продолжить с аккаунтом Instagram'")
}

func fillCredsAndLogin(driver selenium.WebDriver, username, password string) {
	//find with waiting
	var elemUsername selenium.WebElement
	err := driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		foundElem, err := driver.FindElement(selenium.ByCSSSelector, "input[placeholder='Имя пользователя, номер телефона или электронный адрес']")
		if err != nil {
			panic(fmt.Errorf("не удалось найти кнопку 'Имя пользователя, номер телефона или электронный адрес': %v", err))
		}
		elemUsername = foundElem
		visible, err := foundElem.IsDisplayed()
		return visible, err
	}, 10*time.Second)
	if err != nil {
		panic(fmt.Errorf("не удалось найти элемент: %v", err))
	}

	//fill input
	time.Sleep(time.Duration(cryptoRandom(300, 500)) * time.Millisecond)
	err = elemUsername.SendKeys(username)
	if err != nil {
		panic(fmt.Errorf("не удалось ввести 'username': %v", err))
	}

	time.Sleep(1 * time.Second)
	// pageScreenshot(driver, "screen5")

	//find with waiting
	var elemPassword selenium.WebElement
	err = driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		foundElem, err := driver.FindElement(selenium.ByCSSSelector, "input[placeholder='Пароль']")
		if err != nil {
			panic(fmt.Errorf("не удалось найти кнопку 'Пароль': %v", err))
		}
		elemPassword = foundElem
		visible, err := foundElem.IsDisplayed()
		return visible, err
	}, 10*time.Second)
	if err != nil {
		panic(fmt.Errorf("не удалось найти элемент: %v", err))
	}

	//fill input
	time.Sleep(time.Duration(cryptoRandom(300, 500)) * time.Millisecond)
	err = elemPassword.SendKeys(password)
	if err != nil {
		panic(fmt.Errorf("не удалось ввести 'password': %v", err))
	}

	time.Sleep(1 * time.Second)
	// pageScreenshot(driver, "screen6")

	//find with waiting
	var elemSignInButton selenium.WebElement
	err = driver.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		foundElem, err := driver.FindElement(selenium.ByXPATH, "//div[@role='button' and .//div[contains(text(), 'Войти')]]")
		if err != nil {
			panic(fmt.Errorf("не удалось найти кнопку 'Войти': %v", err))
		}
		elemSignInButton = foundElem
		visible, err := foundElem.IsDisplayed()
		return visible, err
	}, 10*time.Second)
	if err != nil {
		panic(fmt.Errorf("не удалось найти элемент: %v", err))
	}

	//click
	time.Sleep(time.Duration(cryptoRandom(300, 500)) * time.Millisecond)
	_, err = driver.ExecuteScript("arguments[0].click();", []interface{}{elemSignInButton})
	if err != nil {
		panic(fmt.Errorf("не удалось кликнуть по кнопке 'Войти': %v", err))
	}
	fmt.Println("Успешно нажали на 'Вход'")
}

// func getAllCookies(driver selenium.WebDriver) {
// 	allCookies, err := driver.GetCookies()
// 	if err != nil {
// 		http.Error(err)
// 	}
// 	fileAllCookies, err := os.Create("allCookies.json")
// 	if err != nil {
// 		http.Error(err)
// 	}
// 	defer fileAllCookies.Close()
// 	encoder := json.NewEncoder(fileAllCookies)
// 	encoder.SetIndent("", "  ")
// 	err = encoder.Encode(allCookies)
// 	if err != nil {
// 		http.Error(err)
// 	}
// 	fmt.Println("Успешно сохранили Cookies в allCookies.json")
// }
