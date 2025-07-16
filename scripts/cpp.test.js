"use strict";

const http = require('http');

function testTranslateAPI() {
    console.log("开始测试翻译API...");

    const requestData = JSON.stringify({
        from: "zh",
        to: "en",
        text: "你好，世界！"
    });

    const options = {
        hostname: 'localhost',
        port: 8989,
        path: '/translate',
        method: 'POST'
    };

    const startTime = Date.now();

    return new Promise((resolve, reject) => {
        const req = http.request(options, (res) => {
            let data = '';

            res.on('data', (chunk) => {
                data += chunk;
            });

            res.on('end', () => {
                const endTime = Date.now();
                const elapsedTime = endTime - startTime;

                try {
                    const response = JSON.parse(data);
                    console.log(`翻译结果: ${response.translatedText}`);
                    console.log(`耗时: ${elapsedTime}ms`);
                    resolve({ result: response.translatedText, time: elapsedTime });
                } catch (error) {
                    console.error(`解析响应失败: ${error.message}`);
                    reject(error);
                }
            });
        });

        req.on('error', (error) => {
            console.error(`请求失败: ${error.message}`);
            reject(error);
        });

        req.write(requestData);
        req.end();
    });
}

// 执行测试
async function runTests() {
    try {
        await testTranslateAPI();
        await testTranslateAPI();
        await testTranslateAPI();
        console.log("所有测试完成");
    } catch (error) {
        console.error("测试失败:", error);
    }
}

runTests();
