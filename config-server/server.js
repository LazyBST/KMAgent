const express = require('express');
const fs = require('fs');
const path = require('path');
const bodyparser = require('body-parser')

const app = express();
app.use(bodyparser.json())
const port = 3000;

app.get('/config', (req, res) => {
    const configFilePath = path.join(__dirname, 'config.json');
    
    fs.readFile(configFilePath, 'utf8', (err, data) => {
        if (err) {
            return res.status(500).send('Error reading config file');
        }
        
        try {
            const config = JSON.parse(data);
            res.json(config);
        } catch (parseErr) {
            res.status(500).send('Error parsing config file');
        }
    });
});

app.post('/status', (req, res) => {
    const data = req.body
    console.log({data})
    return res.status(200).send({
        'status': 'ok'
    });
});

app.listen(port, () => {
    console.log(`Server is running on http://localhost:${port}`);
});