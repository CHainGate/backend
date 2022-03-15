package controller

/*func Register(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	max := big.NewInt(1000000)
	min := big.NewInt(100000)
	verificationCode, err := rand.Int(rand.Reader, max.Sub(max, min))
	if err != nil {
		return
	}
	verificationCode.Add(verificationCode, min)

	emailVerification := model.EmailVerification{
		Email:            data["email"],
		VerificationCode: verificationCode.Uint64(),
	}

	user := model.User{
		Name:              data["name"],
		Email:             data["email"],
		Password:          password,
		IsActive:          false,
		EmailVerification: emailVerification,
	}

	database.DB.Create(&user)

	if user.Id == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	url := utils.Opts.EmailVerificationUrl + "?email=" + user.Email + "&code=" + strconv.FormatUint(user.EmailVerification.VerificationCode, 10)
	err = createVerificationEmail(user.Name, user.Email, url)
	if err != nil {
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	code := r.URL.Query().Get("code")

	var user model.User
	database.DB.Preload("EmailVerification").Where("email = ?", email).First(&user)

	if user.Id == 0 {
		return
	}
	parsedCode, err := strconv.ParseUint(code, 10, 64)
	if err != nil {
		return
	}
	if parsedCode == user.EmailVerification.VerificationCode {
		user.IsActive = true
		database.DB.Save(&user)
	} else {
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var user model.User
	database.DB.Where("email = ?", data["email"]).First(&user)

	if user.Id == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"]))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    strconv.Itoa(int(user.Id)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
	})

	token, err := claims.SignedString([]byte(utils.Opts.JwtSecret))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cookie := &http.Cookie{
		Name:  "token",
		Value: token,
		//Secure: true, HTTPS only on PROD
		HttpOnly: true,
		Expires:  time.Now().Add(time.Hour * 24),
	}
	http.SetCookie(w, cookie)

	err = json.NewEncoder(w).Encode(token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func User(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	token, err := jwt.ParseWithClaims(cookie.Value, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(utils.Opts.JwtSecret), nil
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := token.Claims.(*jwt.RegisteredClaims)

	var user model.User

	database.DB.Where("id = ?", claims.Issuer).First(&user)

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func Logout(w http.ResponseWriter, _ *http.Request) {
	cookie := &http.Cookie{
		Name:  "token",
		Value: "",
		//Secure: true, HTTPS only on PROD
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)
}

func createVerificationEmail(name string, emailTo string, url string) error {
	from := mail.NewEmail("CHaingate", utils.Opts.EmailFrom)
	subject := "Please Verify Your E-Mail"
	to := mail.NewEmail(name, emailTo)
	plainTextContent := "Please Verify your E-Mail: " + url
	htmlContent := "Please Verify your E-Mail: " + url
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(utils.Opts.SendGridApiKey)

	_, err := client.Send(message)
	if err != nil {
		return err
	}
	return nil
}*/

func Test() string {
	return "test"
}
