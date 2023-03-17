package config

///////////////////////////////////////////////////////////////
//
// Using this file to contail multiple web pages.
//
///////////////////////////////////////////////////////////////

func GetPageActiviate(activiate_url string) string {
	return `<!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <title>Activate Your Account</title>
        <style>
            @import url('https://unpkg.com/tailwindcss@latest/dist/tailwind.min.css');

            body {
                font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
                font-size: 14px;
                line-height: 1.4;
                color: #333;
                background-color: #f5f5f5;
            }
    
            .container {
                max-width: 600px;
                margin: 0 auto;
                padding: 20px;
                background-color: #fff;
                border-radius: 5px;
                box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
            }
    
            .logo {
                margin-bottom: 20px;
                text-align: center;
            }
    
            .logo img {
                max-width: 100%;
                height: auto;
            }
    
            .title {
                margin-bottom: 20px;
                text-align: center;
                font-size: 24px;
                font-weight: bold;
            }
    
            .subtitle {
                margin-bottom: 20px;
                text-align: center;
                font-size: 16px;
            }
    
            .button {
                display: inline-block;
                padding: 10px 20px;
                background-color: #5e67df;
                color: #fff;
                text-decoration: none;
                border-radius: 5px;
                transition: background-color 0.3s ease;
            }
    
            .button:hover {
                background-color: #5e67df;
            }
    
            .footer {
                margin-top: 20px;
                text-align: center;
                font-size: 12px;
                color: #999;
            }
        </style>
    </head>
    <body>
    <center class="container">
        <div class="logo">
            <img src="https://via.placeholder.com/150x50" alt="Logo">
        </div>
        <h2 class="text-2xl font-bold mb-4 text-center">X-Ally</h2>

        <div class="title">Activate Your Account</div>
        <div class="subtitle">Thank you for signing up! Please click the button below to activate your account.</div>
        <center><a href="` + activiate_url + `" class="button">Activate Account</a></center>
        <div class="footer">If you did not sign up for this account, please ignore this email.</div>
    </div>
    </body>
    </html>`
}

func GetPageActiviated() string {
	return `<!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <title>Activate Your Account</title>
        <style>
            @import url('https://unpkg.com/tailwindcss@latest/dist/tailwind.min.css');

            body {
                font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;
                font-size: 14px;
                line-height: 1.4;
                color: #333;
                background-color: #f5f5f5;
            }
    
            .container {
                max-width: 600px;
                margin: 0 auto;
                padding: 20px;
                background-color: #fff;
                border-radius: 5px;
                box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
            }
    
            .logo {
                margin-bottom: 20px;
                text-align: center;
            }
    
            .logo img {
                max-width: 100%;
                height: auto;
            }
    
            .title {
                margin-bottom: 20px;
                text-align: center;
                font-size: 24px;
                font-weight: bold;
            }
    
            .subtitle {
                margin-bottom: 20px;
                text-align: center;
                font-size: 16px;
            }
    
            .button {
                display: inline-block;
                padding: 10px 20px;
                background-color: #5e67df;
                color: #fff;
                text-decoration: none;
                border-radius: 5px;
                transition: background-color 0.3s ease;
            }
    
            .button:hover {
                background-color: #3e8e41;
            }
    
            .footer {
                margin-top: 20px;
                text-align: center;
                font-size: 12px;
                color: #999;
            }
        </style>
    </head>
    <body>
    <center class="container">
        <div class="logo">
            <img src="https://via.placeholder.com/150x50" alt="Logo">
        </div>
        <h2 class="text-2xl font-bold mb-4 text-center">X-Ally</h2>

        <div class="title">Congratulations!</div>
        <div class="subtitle">Your account is ready now. Have fun with X-Ally.</div>
        <div class="footer">If you did not sign up for this account, please ignore this email.</div>
    </div>
    </body>
    </html>`
}
