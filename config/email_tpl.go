package config

///////////////////////////////////////////////////////////////
//
// Using this file to contail multiple web pages.
//
///////////////////////////////////////////////////////////////

func GetPageActivate(activate_url string, msg string, title string, tips string) string {
	return `<!DOCTYPE html>
    <html lang="UTF-8">
    <head>
        <meta charset="UTF-8">
        <title>` + title + `</title>
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
            <img src="https://cdn.jsdelivr.net/gh/robinmin/imglanding/images/202303261446199.png" alt="Logo" />
        </div>
        <h2 class="text-2xl font-bold mb-4 text-center">X-Ally</h2>

        <div class="title">` + title + `</div>
        <div class="subtitle">` + msg + `</div>
        <center><a href="` + activate_url + `" class="button">Activate Account</a></center>
        <div class="footer">` + tips + `</div>
    </div>
    </body>
    </html>`
}

func GetPageActiviated(msg string, title string, tips string) string {
	return `<!DOCTYPE html>
    <html lang="UTF-8">
    <head>
        <meta charset="UTF-8">
        <title>` + title + `</title>
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
            <img src="https://cdn.jsdelivr.net/gh/robinmin/imglanding/images/202303261446199.png" alt="Logo">
        </div>
        <h2 class="text-2xl font-bold mb-4 text-center">X-Ally</h2>

        <div class="title">` + title + `</div>
        <div class="subtitle">` + msg + `</div>
        <div class="footer">` + tips + `</div>
    </div>
    </body>
    </html>`
}
