./build/benchmark -grpc-unix /tmp/mtrancore.sock

╔═══════════════════════════════════════════════════════════╗
║       🚀 Translation Service Benchmark Tool 🚀           ║
╚═══════════════════════════════════════════════════════════╝

═══════════════════ Configuration ═════════════════════════
Server URL:    http://localhost:8988
Protocol(s):   all
Model Path:    /models/enzh
Iterations:    100 per test
Concurrency:   1 workers
Warmup:        10 requests
Test Type:     all
═══════════════════════════════════════════════════════════


╔═══════════════════════════════════════════════════════════╗
║  Testing Protocol: HTTP                                     ║
╚═══════════════════════════════════════════════════════════╝

🔌 Establishing HTTP connection...
✅ Connection established successfully!

📦 Loading translation engine...
   Model path: /models/enzh
   Protocol: HTTP
✅ Engine loaded successfully in 1.33ms!

🔥 Warming up with 10 requests...
   Progress: 10/10 - ✅ Completed in 2.68ms (Success: 10/10)

🚀 Starting benchmark tests...
   Test type: all
   Iterations per test: 100
   Concurrency: 1

📋 Running all 10 test cases:

═══════════════════ Test 1/10 ═══════════════════
📊 Running test: Short Greeting
   Text length: 25 chars, HTML: false
   Preview: Hello, how are you today?
   Progress: 100/100 | Success: 100.0% | Avg Latency: 129.67µs   
   Sample translation: 你好,你今天怎么样?

═══════════════════ Test 2/10 ═══════════════════
📊 Running test: News Headline
   Text length: 115 chars, HTML: false
   Preview: Breaking: Scientists discover new approach to renewable energy that could rev...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 321.06µs   
   Sample translation: 打破:科学家发现了可再生能源的新方法,可以彻底改变全球...

═══════════════════ Test 3/10 ═══════════════════
📊 Running test: Product Description
   Text length: 210 chars, HTML: false
   Preview: This premium wireless headphone features active noise cancellation, 30-hour b...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 413.74µs   
   Sample translation: 这款高级无线耳机具有主动降噪功能,30小时电池续航时间�..

═══════════════════ Test 4/10 ═══════════════════
📊 Running test: Email Message
   Text length: 324 chars, HTML: false
   Preview: Dear Team, I hope this message finds you well. I wanted to follow up on our d...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 515.08µs   
   Sample translation: 亲爱的团队,我希望这个消息能很好地找到你。 我想从昨天...

═══════════════════ Test 5/10 ═══════════════════
📊 Running test: Technical Article
   Text length: 518 chars, HTML: false
   Preview: Machine learning models require large amounts of training data to achieve opt...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 844.56µs   
   Sample translation: 机器学习模型需要大量的训练数据才能达到最佳性能。 该�..

═══════════════════ Test 6/10 ═══════════════════
📊 Running test: Legal Notice
   Text length: 352 chars, HTML: false
   Preview: By accessing this website, you agree to be bound by these terms and condition...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 1.08ms   
   Sample translation: 访问本网站即表示您同意受这些条款和条件的约束。 本公�..

═══════════════════ Test 7/10 ═══════════════════
📊 Running test: HTML Article
   Text length: 390 chars, HTML: true
   Preview: <article><h1>Welcome to Modern Web Development</h1><p>Learn the latest techno...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 1.27ms   
   Sample translation: <article><h1>欢迎来到现代Web开发</h1><p><strong>了解构建令人惊...

═══════════════════ Test 8/10 ═══════════════════
📊 Running test: Medical Information
   Text length: 356 chars, HTML: false
   Preview: Patient care requires comprehensive assessment and individualized treatment p...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 1.67ms   
   Sample translation: 患者护理需要全面评估和个性化治疗规划。 医疗保健提供�..

═══════════════════ Test 9/10 ═══════════════════
📊 Running test: Customer Support
   Text length: 352 chars, HTML: false
   Preview: Thank you for contacting our support team. We understand you're experiencing ...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 618.00µs   
   Sample translation: 感谢您联系我们的支持团队。 我们了解您在登录帐户时遇�..

═══════════════════ Test 10/10 ═══════════════════
📊 Running test: Long Document
   Text length: 1197 chars, HTML: false
   Preview: The rapid advancement of technology has fundamentally transformed how busines...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 2.68ms   
   Sample translation: 技术的快速发展从根本上改变了企业在现代经济中的运作�..


✅ All tests completed for HTTP protocol!
┌───────────────────────────────────────────────────────────┐
│ Test: Short Greeting                                      │
│ Protocol: HTTP                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    13.04ms                          │
│   Throughput:        7670.47                      req/s │
│   Avg Time/Request:  130.37µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               82.64µs                          │
│   Average:           129.67µs                         │
│   Median (P50):      114.93µs                         │
│   P90:               196.59µs                         │
│   P95:               224.46µs                         │
│   P99:               408.05µs                         │
│   Max:               408.05µs                         │
│   Spread (Max-Min):  325.41µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.95x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: News Headline                                       │
│ Protocol: HTTP                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    32.20ms                          │
│   Throughput:        3105.87                      req/s │
│   Avg Time/Request:  321.97µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               197.18µs                         │
│   Average:           321.06µs                         │
│   Median (P50):      350.79µs                         │
│   P90:               420.54µs                         │
│   P95:               456.05µs                         │
│   P99:               606.80µs                         │
│   Max:               606.80µs                         │
│   Spread (Max-Min):  409.62µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.30x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Product Description                                 │
│ Protocol: HTTP                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    41.45ms                          │
│   Throughput:        2412.39                      req/s │
│   Avg Time/Request:  414.53µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               315.62µs                         │
│   Average:           413.74µs                         │
│   Median (P50):      365.76µs                         │
│   P90:               558.52µs                         │
│   P95:               594.27µs                         │
│   P99:               1.41ms                           │
│   Max:               1.41ms                           │
│   Spread (Max-Min):  1.09ms                           │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.62x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Email Message                                       │
│ Protocol: HTTP                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    51.59ms                          │
│   Throughput:        1938.26                      req/s │
│   Avg Time/Request:  515.92µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               428.98µs                         │
│   Average:           515.08µs                         │
│   Median (P50):      488.40µs                         │
│   P90:               601.23µs                         │
│   P95:               682.60µs                         │
│   P99:               1.02ms                           │
│   Max:               1.02ms                           │
│   Spread (Max-Min):  588.89µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.40x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Technical Article                                   │
│ Protocol: HTTP                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    84.55ms                          │
│   Throughput:        1182.78                      req/s │
│   Avg Time/Request:  845.47µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               697.58µs                         │
│   Average:           844.56µs                         │
│   Median (P50):      777.15µs                         │
│   P90:               1.13ms                           │
│   P95:               1.18ms                           │
│   P99:               1.33ms                           │
│   Max:               1.33ms                           │
│   Spread (Max-Min):  629.68µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.51x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Legal Notice                                        │
│ Protocol: HTTP                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    107.96ms                         │
│   Throughput:        926.29                       req/s │
│   Avg Time/Request:  1.08ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               460.86µs                         │
│   Average:           1.08ms                           │
│   Median (P50):      572.32µs                         │
│   P90:               846.04µs                         │
│   P95:               941.16µs                         │
│   P99:               46.45ms                          │
│   Max:               46.45ms                          │
│   Spread (Max-Min):  45.99ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.64x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: HTML Article                                        │
│ Protocol: HTTP                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    127.17ms                         │
│   Throughput:        786.31                       req/s │
│   Avg Time/Request:  1.27ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               611.75µs                         │
│   Average:           1.27ms                           │
│   Median (P50):      678.21µs                         │
│   P90:               941.84µs                         │
│   P95:               981.30µs                         │
│   P99:               54.48ms                          │
│   Max:               54.48ms                          │
│   Spread (Max-Min):  53.87ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.45x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Medical Information                                 │
│ Protocol: HTTP                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    166.75ms                         │
│   Throughput:        599.69                       req/s │
│   Avg Time/Request:  1.67ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               530.00µs                         │
│   Average:           1.67ms                           │
│   Median (P50):      608.20µs                         │
│   P90:               939.79µs                         │
│   P95:               1.03ms                           │
│   P99:               98.64ms                          │
│   Max:               98.64ms                          │
│   Spread (Max-Min):  98.11ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.70x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Customer Support                                    │
│ Protocol: HTTP                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    61.91ms                          │
│   Throughput:        1615.21                      req/s │
│   Avg Time/Request:  619.11µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               468.93µs                         │
│   Average:           618.00µs                         │
│   Median (P50):      547.41µs                         │
│   P90:               847.57µs                         │
│   P95:               915.05µs                         │
│   P99:               1.02ms                           │
│   Max:               1.02ms                           │
│   Spread (Max-Min):  548.58µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.67x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Long Document                                       │
│ Protocol: HTTP                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    268.12ms                         │
│   Throughput:        372.97                       req/s │
│   Avg Time/Request:  2.68ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               1.54ms                           │
│   Average:           2.68ms                           │
│   Median (P50):      1.65ms                           │
│   P90:               1.92ms                           │
│   P95:               2.06ms                           │
│   P99:               99.42ms                          │
│   Max:               99.42ms                          │
│   Spread (Max-Min):  97.87ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.25x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘


╔═══════════════════════════════════════════════════════════╗
║  Testing Protocol: GRPC                                     ║
╚═══════════════════════════════════════════════════════════╝

🔌 Establishing GRPC connection...
✅ Connection established successfully!

📦 Loading translation engine...
   Model path: /models/enzh
   Protocol: GRPC
✅ Engine loaded successfully in 1.35ms!

🔥 Warming up with 10 requests...
   Progress: 10/10 - ✅ Completed in 2.20ms (Success: 10/10)

🚀 Starting benchmark tests...
   Test type: all
   Iterations per test: 100
   Concurrency: 1

📋 Running all 10 test cases:

═══════════════════ Test 1/10 ═══════════════════
📊 Running test: Short Greeting
   Text length: 25 chars, HTML: false
   Preview: Hello, how are you today?
   Progress: 100/100 | Success: 100.0% | Avg Latency: 173.00µs   
   Sample translation: 你好,你今天怎么样?

═══════════════════ Test 2/10 ═══════════════════
📊 Running test: News Headline
   Text length: 115 chars, HTML: false
   Preview: Breaking: Scientists discover new approach to renewable energy that could rev...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 354.09µs   
   Sample translation: 打破:科学家发现了可再生能源的新方法,可以彻底改变全球...

═══════════════════ Test 3/10 ═══════════════════
📊 Running test: Product Description
   Text length: 210 chars, HTML: false
   Preview: This premium wireless headphone features active noise cancellation, 30-hour b...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 428.38µs   
   Sample translation: 这款高级无线耳机具有主动降噪功能,30小时电池续航时间�..

═══════════════════ Test 4/10 ═══════════════════
📊 Running test: Email Message
   Text length: 324 chars, HTML: false
   Preview: Dear Team, I hope this message finds you well. I wanted to follow up on our d...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 577.43µs   
   Sample translation: 亲爱的团队,我希望这个消息能很好地找到你。 我想从昨天...

═══════════════════ Test 5/10 ═══════════════════
📊 Running test: Technical Article
   Text length: 518 chars, HTML: false
   Preview: Machine learning models require large amounts of training data to achieve opt...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 875.57µs   
   Sample translation: 机器学习模型需要大量的训练数据才能达到最佳性能。 该�..

═══════════════════ Test 6/10 ═══════════════════
📊 Running test: Legal Notice
   Text length: 352 chars, HTML: false
   Preview: By accessing this website, you agree to be bound by these terms and condition...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 1.08ms   
   Sample translation: 访问本网站即表示您同意受这些条款和条件的约束。 本公�..

═══════════════════ Test 7/10 ═══════════════════
📊 Running test: HTML Article
   Text length: 390 chars, HTML: true
   Preview: <article><h1>Welcome to Modern Web Development</h1><p>Learn the latest techno...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 1.26ms   
   Sample translation: <article><h1>欢迎来到现代Web开发</h1><p><strong>了解构建令人惊...

═══════════════════ Test 8/10 ═══════════════════
📊 Running test: Medical Information
   Text length: 356 chars, HTML: false
   Preview: Patient care requires comprehensive assessment and individualized treatment p...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 1.64ms   
   Sample translation: 患者护理需要全面评估和个性化治疗规划。 医疗保健提供�..

═══════════════════ Test 9/10 ═══════════════════
📊 Running test: Customer Support
   Text length: 352 chars, HTML: false
   Preview: Thank you for contacting our support team. We understand you're experiencing ...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 672.37µs   
   Sample translation: 感谢您联系我们的支持团队。 我们了解您在登录帐户时遇�..

═══════════════════ Test 10/10 ═══════════════════
📊 Running test: Long Document
   Text length: 1197 chars, HTML: false
   Preview: The rapid advancement of technology has fundamentally transformed how busines...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 2.84ms   
   Sample translation: 技术的快速发展从根本上改变了企业在现代经济中的运作�..


✅ All tests completed for GRPC protocol!
┌───────────────────────────────────────────────────────────┐
│ Test: Short Greeting                                      │
│ Protocol: GRPC                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    17.39ms                          │
│   Throughput:        5750.81                      req/s │
│   Avg Time/Request:  173.89µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               99.09µs                          │
│   Average:           173.00µs                         │
│   Median (P50):      183.06µs                         │
│   P90:               233.30µs                         │
│   P95:               261.59µs                         │
│   P99:               277.08µs                         │
│   Max:               277.08µs                         │
│   Spread (Max-Min):  177.99µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.43x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: News Headline                                       │
│ Protocol: GRPC                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    35.51ms                          │
│   Throughput:        2816.31                      req/s │
│   Avg Time/Request:  355.07µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               221.75µs                         │
│   Average:           354.09µs                         │
│   Median (P50):      374.68µs                         │
│   P90:               402.05µs                         │
│   P95:               426.02µs                         │
│   P99:               506.93µs                         │
│   Max:               506.93µs                         │
│   Spread (Max-Min):  285.17µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Excellent                (1.14x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Product Description                                 │
│ Protocol: GRPC                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    42.92ms                          │
│   Throughput:        2329.67                      req/s │
│   Avg Time/Request:  429.25µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               336.26µs                         │
│   Average:           428.38µs                         │
│   Median (P50):      400.59µs                         │
│   P90:               550.14µs                         │
│   P95:               591.65µs                         │
│   P99:               741.14µs                         │
│   Max:               741.14µs                         │
│   Spread (Max-Min):  404.88µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.48x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Email Message                                       │
│ Protocol: GRPC                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    57.84ms                          │
│   Throughput:        1728.75                      req/s │
│   Avg Time/Request:  578.45µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               452.41µs                         │
│   Average:           577.43µs                         │
│   Median (P50):      532.93µs                         │
│   P90:               770.10µs                         │
│   P95:               799.97µs                         │
│   P99:               1.03ms                           │
│   Max:               1.03ms                           │
│   Spread (Max-Min):  578.84µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.50x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Technical Article                                   │
│ Protocol: GRPC                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    87.67ms                          │
│   Throughput:        1140.66                      req/s │
│   Avg Time/Request:  876.69µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               739.12µs                         │
│   Average:           875.57µs                         │
│   Median (P50):      802.67µs                         │
│   P90:               1.18ms                           │
│   P95:               1.25ms                           │
│   P99:               1.92ms                           │
│   Max:               1.92ms                           │
│   Spread (Max-Min):  1.18ms                           │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.56x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Legal Notice                                        │
│ Protocol: GRPC                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    108.15ms                         │
│   Throughput:        924.63                       req/s │
│   Avg Time/Request:  1.08ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               492.99µs                         │
│   Average:           1.08ms                           │
│   Median (P50):      569.46µs                         │
│   P90:               801.13µs                         │
│   P95:               869.05µs                         │
│   P99:               46.83ms                          │
│   Max:               46.83ms                          │
│   Spread (Max-Min):  46.34ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.53x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: HTML Article                                        │
│ Protocol: GRPC                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    126.02ms                         │
│   Throughput:        793.54                       req/s │
│   Avg Time/Request:  1.26ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               631.78µs                         │
│   Average:           1.26ms                           │
│   Median (P50):      685.20µs                         │
│   P90:               795.52µs                         │
│   P95:               837.23µs                         │
│   P99:               56.56ms                          │
│   Max:               56.56ms                          │
│   Spread (Max-Min):  55.93ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.22x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Medical Information                                 │
│ Protocol: GRPC                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    163.69ms                         │
│   Throughput:        610.93                       req/s │
│   Avg Time/Request:  1.64ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               565.93µs                         │
│   Average:           1.64ms                           │
│   Median (P50):      630.49µs                         │
│   P90:               755.71µs                         │
│   P95:               875.90µs                         │
│   P99:               98.91ms                          │
│   Max:               98.91ms                          │
│   Spread (Max-Min):  98.34ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.39x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Customer Support                                    │
│ Protocol: GRPC                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    67.37ms                          │
│   Throughput:        1484.31                      req/s │
│   Avg Time/Request:  673.71µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               491.50µs                         │
│   Average:           672.37µs                         │
│   Median (P50):      614.07µs                         │
│   P90:               869.26µs                         │
│   P95:               881.21µs                         │
│   P99:               1.18ms                           │
│   Max:               1.18ms                           │
│   Spread (Max-Min):  686.56µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.44x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Long Document                                       │
│ Protocol: GRPC                                            │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    284.47ms                         │
│   Throughput:        351.53                       req/s │
│   Avg Time/Request:  2.84ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               1.61ms                           │
│   Average:           2.84ms                           │
│   Median (P50):      1.74ms                           │
│   P90:               2.48ms                           │
│   P95:               2.56ms                           │
│   P99:               98.58ms                          │
│   Max:               98.58ms                          │
│   Spread (Max-Min):  96.97ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.47x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘


╔═══════════════════════════════════════════════════════════╗
║  Testing Protocol: WS                                       ║
╚═══════════════════════════════════════════════════════════╝

🔌 Establishing WS connection...
✅ Connection established successfully!
📡 WebSocket pool size: 1 connections

📦 Loading translation engine...
   Model path: /models/enzh
   Protocol: WS
✅ Engine loaded successfully in 122.88µs!

🔥 Warming up with 10 requests...
   Progress: 10/10 - ✅ Completed in 1.46ms (Success: 10/10)

🚀 Starting benchmark tests...
   Test type: all
   Iterations per test: 100
   Concurrency: 1

📋 Running all 10 test cases:

═══════════════════ Test 1/10 ═══════════════════
📊 Running test: Short Greeting
   Text length: 25 chars, HTML: false
   Preview: Hello, how are you today?
   Progress: 100/100 | Success: 100.0% | Avg Latency: 79.98µs   
   Sample translation: 你好,你今天怎么样?

═══════════════════ Test 2/10 ═══════════════════
📊 Running test: News Headline
   Text length: 115 chars, HTML: false
   Preview: Breaking: Scientists discover new approach to renewable energy that could rev...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 263.55µs   
   Sample translation: 打破:科学家发现了可再生能源的新方法,可以彻底改变全球...

═══════════════════ Test 3/10 ═══════════════════
📊 Running test: Product Description
   Text length: 210 chars, HTML: false
   Preview: This premium wireless headphone features active noise cancellation, 30-hour b...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 331.40µs   
   Sample translation: 这款高级无线耳机具有主动降噪功能,30小时电池续航时间�..

═══════════════════ Test 4/10 ═══════════════════
📊 Running test: Email Message
   Text length: 324 chars, HTML: false
   Preview: Dear Team, I hope this message finds you well. I wanted to follow up on our d...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 449.97µs   
   Sample translation: 亲爱的团队,我希望这个消息能很好地找到你。 我想从昨天...

═══════════════════ Test 5/10 ═══════════════════
📊 Running test: Technical Article
   Text length: 518 chars, HTML: false
   Preview: Machine learning models require large amounts of training data to achieve opt...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 736.60µs   
   Sample translation: 机器学习模型需要大量的训练数据才能达到最佳性能。 该�..

═══════════════════ Test 6/10 ═══════════════════
📊 Running test: Legal Notice
   Text length: 352 chars, HTML: false
   Preview: By accessing this website, you agree to be bound by these terms and condition...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 941.98µs   
   Sample translation: 访问本网站即表示您同意受这些条款和条件的约束。 本公�..

═══════════════════ Test 7/10 ═══════════════════
📊 Running test: HTML Article
   Text length: 390 chars, HTML: true
   Preview: <article><h1>Welcome to Modern Web Development</h1><p>Learn the latest techno...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 1.19ms   
   Sample translation: <article><h1>欢迎来到现代Web开发</h1><p><strong>了解构建令人惊...

═══════════════════ Test 8/10 ═══════════════════
📊 Running test: Medical Information
   Text length: 356 chars, HTML: false
   Preview: Patient care requires comprehensive assessment and individualized treatment p...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 1.55ms   
   Sample translation: 患者护理需要全面评估和个性化治疗规划。 医疗保健提供�..

═══════════════════ Test 9/10 ═══════════════════
📊 Running test: Customer Support
   Text length: 352 chars, HTML: false
   Preview: Thank you for contacting our support team. We understand you're experiencing ...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 500.29µs   
   Sample translation: 感谢您联系我们的支持团队。 我们了解您在登录帐户时遇�..

═══════════════════ Test 10/10 ═══════════════════
📊 Running test: Long Document
   Text length: 1197 chars, HTML: false
   Preview: The rapid advancement of technology has fundamentally transformed how busines...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 2.63ms   
   Sample translation: 技术的快速发展从根本上改变了企业在现代经济中的运作�..


✅ All tests completed for WS protocol!
┌───────────────────────────────────────────────────────────┐
│ Test: Short Greeting                                      │
│ Protocol: WS                                              │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    8.04ms                           │
│   Throughput:        12445.50                     req/s │
│   Avg Time/Request:  80.35µs                          │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               70.32µs                          │
│   Average:           79.98µs                          │
│   Median (P50):      74.03µs                          │
│   P90:               103.83µs                         │
│   P95:               119.06µs                         │
│   P99:               145.55µs                         │
│   Max:               145.55µs                         │
│   Spread (Max-Min):  75.23µs                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.61x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: News Headline                                       │
│ Protocol: WS                                              │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    26.43ms                          │
│   Throughput:        3783.75                      req/s │
│   Avg Time/Request:  264.29µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               187.70µs                         │
│   Average:           263.55µs                         │
│   Median (P50):      246.26µs                         │
│   P90:               349.41µs                         │
│   P95:               377.36µs                         │
│   P99:               492.99µs                         │
│   Max:               492.99µs                         │
│   Spread (Max-Min):  305.29µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.53x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Product Description                                 │
│ Protocol: WS                                              │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    33.19ms                          │
│   Throughput:        3013.37                      req/s │
│   Avg Time/Request:  331.85µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               295.66µs                         │
│   Average:           331.40µs                         │
│   Median (P50):      311.80µs                         │
│   P90:               355.96µs                         │
│   P95:               565.38µs                         │
│   P99:               656.79µs                         │
│   Max:               656.79µs                         │
│   Spread (Max-Min):  361.12µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.81x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Email Message                                       │
│ Protocol: WS                                              │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    45.05ms                          │
│   Throughput:        2219.82                      req/s │
│   Avg Time/Request:  450.49µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               406.45µs                         │
│   Average:           449.97µs                         │
│   Median (P50):      433.33µs                         │
│   P90:               489.13µs                         │
│   P95:               540.31µs                         │
│   P99:               841.58µs                         │
│   Max:               841.58µs                         │
│   Spread (Max-Min):  435.13µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.25x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Technical Article                                   │
│ Protocol: WS                                              │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    73.72ms                          │
│   Throughput:        1356.38                      req/s │
│   Avg Time/Request:  737.26µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               685.02µs                         │
│   Average:           736.60µs                         │
│   Median (P50):      710.86µs                         │
│   P90:               829.60µs                         │
│   P95:               923.98µs                         │
│   P99:               1.08ms                           │
│   Max:               1.08ms                           │
│   Spread (Max-Min):  395.71µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.30x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Legal Notice                                        │
│ Protocol: WS                                              │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    94.27ms                          │
│   Throughput:        1060.74                      req/s │
│   Avg Time/Request:  942.74µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               435.48µs                         │
│   Average:           941.98µs                         │
│   Median (P50):      462.08µs                         │
│   P90:               554.22µs                         │
│   P95:               595.92µs                         │
│   P99:               46.96ms                          │
│   Max:               46.96ms                          │
│   Spread (Max-Min):  46.53ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.29x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: HTML Article                                        │
│ Protocol: WS                                              │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    119.03ms                         │
│   Throughput:        840.13                       req/s │
│   Avg Time/Request:  1.19ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               600.85µs                         │
│   Average:           1.19ms                           │
│   Median (P50):      632.47µs                         │
│   P90:               746.78µs                         │
│   P95:               813.02µs                         │
│   P99:               54.12ms                          │
│   Max:               54.12ms                          │
│   Spread (Max-Min):  53.52ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.29x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Medical Information                                 │
│ Protocol: WS                                              │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    155.01ms                         │
│   Throughput:        645.10                       req/s │
│   Avg Time/Request:  1.55ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               512.44µs                         │
│   Average:           1.55ms                           │
│   Median (P50):      538.71µs                         │
│   P90:               681.92µs                         │
│   P95:               797.57µs                         │
│   P99:               98.71ms                          │
│   Max:               98.71ms                          │
│   Spread (Max-Min):  98.20ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.48x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Customer Support                                    │
│ Protocol: WS                                              │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    50.09ms                          │
│   Throughput:        1996.43                      req/s │
│   Avg Time/Request:  500.89µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               453.06µs                         │
│   Average:           500.29µs                         │
│   Median (P50):      482.22µs                         │
│   P90:               549.04µs                         │
│   P95:               674.14µs                         │
│   P99:               894.17µs                         │
│   Max:               894.17µs                         │
│   Spread (Max-Min):  441.11µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.40x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Long Document                                       │
│ Protocol: WS                                              │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    263.05ms                         │
│   Throughput:        380.15                       req/s │
│   Avg Time/Request:  2.63ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               1.55ms                           │
│   Average:           2.63ms                           │
│   Median (P50):      1.62ms                           │
│   P90:               2.01ms                           │
│   P95:               2.07ms                           │
│   P99:               97.28ms                          │
│   Max:               97.28ms                          │
│   Spread (Max-Min):  95.74ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.28x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘


╔═══════════════════════════════════════════════════════════╗
║  Testing Protocol: GRPC-UNIX                                ║
╚═══════════════════════════════════════════════════════════╝

🔌 Establishing GRPC-UNIX connection...
✅ Connection established successfully!

📦 Loading translation engine...
   Model path: /models/enzh
   Protocol: GRPC-UNIX
✅ Engine loaded successfully in 598.07µs!

🔥 Warming up with 10 requests...
   Progress: 10/10 - ✅ Completed in 1.74ms (Success: 10/10)

🚀 Starting benchmark tests...
   Test type: all
   Iterations per test: 100
   Concurrency: 1

📋 Running all 10 test cases:

═══════════════════ Test 1/10 ═══════════════════
📊 Running test: Short Greeting
   Text length: 25 chars, HTML: false
   Preview: Hello, how are you today?
   Progress: 100/100 | Success: 100.0% | Avg Latency: 129.42µs   
   Sample translation: 你好,你今天怎么样?

═══════════════════ Test 2/10 ═══════════════════
📊 Running test: News Headline
   Text length: 115 chars, HTML: false
   Preview: Breaking: Scientists discover new approach to renewable energy that could rev...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 361.54µs   
   Sample translation: 打破:科学家发现了可再生能源的新方法,可以彻底改变全球...

═══════════════════ Test 3/10 ═══════════════════
📊 Running test: Product Description
   Text length: 210 chars, HTML: false
   Preview: This premium wireless headphone features active noise cancellation, 30-hour b...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 446.85µs   
   Sample translation: 这款高级无线耳机具有主动降噪功能,30小时电池续航时间�..

═══════════════════ Test 4/10 ═══════════════════
📊 Running test: Email Message
   Text length: 324 chars, HTML: false
   Preview: Dear Team, I hope this message finds you well. I wanted to follow up on our d...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 608.44µs   
   Sample translation: 亲爱的团队,我希望这个消息能很好地找到你。 我想从昨天...

═══════════════════ Test 5/10 ═══════════════════
📊 Running test: Technical Article
   Text length: 518 chars, HTML: false
   Preview: Machine learning models require large amounts of training data to achieve opt...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 847.72µs   
   Sample translation: 机器学习模型需要大量的训练数据才能达到最佳性能。 该�..

═══════════════════ Test 6/10 ═══════════════════
📊 Running test: Legal Notice
   Text length: 352 chars, HTML: false
   Preview: By accessing this website, you agree to be bound by these terms and condition...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 1.03ms   
   Sample translation: 访问本网站即表示您同意受这些条款和条件的约束。 本公�..

═══════════════════ Test 7/10 ═══════════════════
📊 Running test: HTML Article
   Text length: 390 chars, HTML: true
   Preview: <article><h1>Welcome to Modern Web Development</h1><p>Learn the latest techno...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 1.32ms   
   Sample translation: <article><h1>欢迎来到现代Web开发</h1><p><strong>了解构建令人惊...

═══════════════════ Test 8/10 ═══════════════════
📊 Running test: Medical Information
   Text length: 356 chars, HTML: false
   Preview: Patient care requires comprehensive assessment and individualized treatment p...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 1.68ms   
   Sample translation: 患者护理需要全面评估和个性化治疗规划。 医疗保健提供�..

═══════════════════ Test 9/10 ═══════════════════
📊 Running test: Customer Support
   Text length: 352 chars, HTML: false
   Preview: Thank you for contacting our support team. We understand you're experiencing ...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 576.42µs   
   Sample translation: 感谢您联系我们的支持团队。 我们了解您在登录帐户时遇�..

═══════════════════ Test 10/10 ═══════════════════
📊 Running test: Long Document
   Text length: 1197 chars, HTML: false
   Preview: The rapid advancement of technology has fundamentally transformed how busines...
   Progress: 100/100 | Success: 100.0% | Avg Latency: 2.86ms   
   Sample translation: 技术的快速发展从根本上改变了企业在现代经济中的运作�..


✅ All tests completed for GRPC-UNIX protocol!
┌───────────────────────────────────────────────────────────┐
│ Test: Short Greeting                                      │
│ Protocol: GRPC-UNIX                                       │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    13.00ms                          │
│   Throughput:        7691.49                      req/s │
│   Avg Time/Request:  130.01µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               88.23µs                          │
│   Average:           129.42µs                         │
│   Median (P50):      106.68µs                         │
│   P90:               209.51µs                         │
│   P95:               232.04µs                         │
│   P99:               265.57µs                         │
│   Max:               265.57µs                         │
│   Spread (Max-Min):  177.33µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Poor                     (2.18x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: News Headline                                       │
│ Protocol: GRPC-UNIX                                       │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    36.28ms                          │
│   Throughput:        2756.17                      req/s │
│   Avg Time/Request:  362.82µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               242.54µs                         │
│   Average:           361.54µs                         │
│   Median (P50):      365.96µs                         │
│   P90:               403.85µs                         │
│   P95:               422.44µs                         │
│   P99:               552.15µs                         │
│   Max:               552.15µs                         │
│   Spread (Max-Min):  309.61µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Excellent                (1.15x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Product Description                                 │
│ Protocol: GRPC-UNIX                                       │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    44.77ms                          │
│   Throughput:        2233.62                      req/s │
│   Avg Time/Request:  447.70µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               326.54µs                         │
│   Average:           446.85µs                         │
│   Median (P50):      416.86µs                         │
│   P90:               557.06µs                         │
│   P95:               583.48µs                         │
│   P99:               1.09ms                           │
│   Max:               1.09ms                           │
│   Spread (Max-Min):  766.05µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.40x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Email Message                                       │
│ Protocol: GRPC-UNIX                                       │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    60.94ms                          │
│   Throughput:        1641.08                      req/s │
│   Avg Time/Request:  609.36µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               450.62µs                         │
│   Average:           608.44µs                         │
│   Median (P50):      575.40µs                         │
│   P90:               761.24µs                         │
│   P95:               792.05µs                         │
│   P99:               903.55µs                         │
│   Max:               903.55µs                         │
│   Spread (Max-Min):  452.93µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.38x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Technical Article                                   │
│ Protocol: GRPC-UNIX                                       │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    84.88ms                          │
│   Throughput:        1178.18                      req/s │
│   Avg Time/Request:  848.77µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               712.41µs                         │
│   Average:           847.72µs                         │
│   Median (P50):      786.66µs                         │
│   P90:               1.12ms                           │
│   P95:               1.17ms                           │
│   P99:               1.30ms                           │
│   Max:               1.30ms                           │
│   Spread (Max-Min):  585.33µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.49x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Legal Notice                                        │
│ Protocol: GRPC-UNIX                                       │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    103.44ms                         │
│   Throughput:        966.71                       req/s │
│   Avg Time/Request:  1.03ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               455.05µs                         │
│   Average:           1.03ms                           │
│   Median (P50):      516.43µs                         │
│   P90:               778.65µs                         │
│   P95:               827.58µs                         │
│   P99:               46.62ms                          │
│   Max:               46.62ms                          │
│   Spread (Max-Min):  46.17ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.60x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: HTML Article                                        │
│ Protocol: GRPC-UNIX                                       │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    132.20ms                         │
│   Throughput:        756.43                       req/s │
│   Avg Time/Request:  1.32ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               612.16µs                         │
│   Average:           1.32ms                           │
│   Median (P50):      704.49µs                         │
│   P90:               978.84µs                         │
│   P95:               1.10ms                           │
│   P99:               55.70ms                          │
│   Max:               55.70ms                          │
│   Spread (Max-Min):  55.09ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.56x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Medical Information                                 │
│ Protocol: GRPC-UNIX                                       │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    167.69ms                         │
│   Throughput:        596.33                       req/s │
│   Avg Time/Request:  1.68ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               523.30µs                         │
│   Average:           1.68ms                           │
│   Median (P50):      614.89µs                         │
│   P90:               933.54µs                         │
│   P95:               983.52µs                         │
│   P99:               99.69ms                          │
│   Max:               99.69ms                          │
│   Spread (Max-Min):  99.17ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.60x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Customer Support                                    │
│ Protocol: GRPC-UNIX                                       │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    57.73ms                          │
│   Throughput:        1732.21                      req/s │
│   Avg Time/Request:  577.30µs                         │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               471.47µs                         │
│   Average:           576.42µs                         │
│   Median (P50):      543.52µs                         │
│   P90:               759.84µs                         │
│   P95:               807.95µs                         │
│   P99:               950.28µs                         │
│   Max:               950.28µs                         │
│   Spread (Max-Min):  478.81µs                         │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Good                     (1.49x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│ Test: Long Document                                       │
│ Protocol: GRPC-UNIX                                       │
├───────────────────────────────────────────────────────────┤
│ 📊 Request Statistics:                                   │
│   Total Requests:    100                              │
│   ✅ Successful:     100                              │
│   ❌ Failed:         0                                │
│   Success Rate:      100.00                        % │
│                                                           │
│ ⏱️  Timing Statistics:                                    │
│   Total Duration:    286.24ms                         │
│   Throughput:        349.35                       req/s │
│   Avg Time/Request:  2.86ms                           │
│                                                           │
│ 📈 Latency Distribution:                                 │
│   Min:               1.58ms                           │
│   Average:           2.86ms                           │
│   Median (P50):      1.75ms                           │
│   P90:               2.49ms                           │
│   P95:               2.75ms                           │
│   P99:               96.38ms                          │
│   Max:               96.38ms                          │
│   Spread (Max-Min):  94.80ms                          │
│                                                           │
│ 🎯 Performance Indicators:                               │
│   Consistency:       Fair                     (1.57x) │
│   Throughput Class:  Excellent                        │
└───────────────────────────────────────────────────────────┘



╔═══════════════════════════════════════════════════════════╗
║              📊 Protocol Comparison Summary               ║
╚═══════════════════════════════════════════════════════════╝

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Test 1: Email Message
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │  Success │   Throughput │        Avg │        P50 │        P95 │        P99
─────────────┼──────────┼──────────────┼────────────┼────────────┼────────────┼──────────
GRPC         │  100.0% │    1728.75/s │   577.43µs │   532.93µs │   799.97µs │     1.03ms
GRPC-UNIX    │  100.0% │    1641.08/s │   608.44µs │   575.40µs │   792.05µs │   903.55µs
HTTP         │  100.0% │    1938.26/s │   515.08µs │   488.40µs │   682.60µs │     1.02ms
WS           │  100.0% │    2219.82/s │   449.97µs │   433.33µs │   540.31µs │   841.58µs 🏆

Performance Analysis:
  • WS is 35.3% faster than GRPC-UNIX in throughput
  • WS has the best P95 latency: 540.31µs

Legend: 🏆 Best Overall | ⚡ Highest Throughput | 🎯 Lowest Latency

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Test 2: Legal Notice
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │  Success │   Throughput │        Avg │        P50 │        P95 │        P99
─────────────┼──────────┼──────────────┼────────────┼────────────┼────────────┼──────────
GRPC         │  100.0% │     924.63/s │     1.08ms │   569.46µs │   869.05µs │    46.83ms
GRPC-UNIX    │  100.0% │     966.71/s │     1.03ms │   516.43µs │   827.58µs │    46.62ms
HTTP         │  100.0% │     926.29/s │     1.08ms │   572.32µs │   941.16µs │    46.45ms
WS           │  100.0% │    1060.74/s │   941.98µs │   462.08µs │   595.92µs │    46.96ms 🏆

Performance Analysis:
  • WS is 14.7% faster than GRPC in throughput
  • WS has the best P95 latency: 595.92µs

Legend: 🏆 Best Overall | ⚡ Highest Throughput | 🎯 Lowest Latency

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Test 3: HTML Article
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │  Success │   Throughput │        Avg │        P50 │        P95 │        P99
─────────────┼──────────┼──────────────┼────────────┼────────────┼────────────┼──────────
GRPC         │  100.0% │     793.54/s │     1.26ms │   685.20µs │   837.23µs │    56.56ms
GRPC-UNIX    │  100.0% │     756.43/s │     1.32ms │   704.49µs │     1.10ms │    55.70ms
HTTP         │  100.0% │     786.31/s │     1.27ms │   678.21µs │   981.30µs │    54.48ms
WS           │  100.0% │     840.13/s │     1.19ms │   632.47µs │   813.02µs │    54.12ms 🏆

Performance Analysis:
  • WS is 11.1% faster than GRPC-UNIX in throughput
  • WS has the best P95 latency: 813.02µs

Legend: 🏆 Best Overall | ⚡ Highest Throughput | 🎯 Lowest Latency

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Test 4: Medical Information
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │  Success │   Throughput │        Avg │        P50 │        P95 │        P99
─────────────┼──────────┼──────────────┼────────────┼────────────┼────────────┼──────────
GRPC         │  100.0% │     610.93/s │     1.64ms │   630.49µs │   875.90µs │    98.91ms
GRPC-UNIX    │  100.0% │     596.33/s │     1.68ms │   614.89µs │   983.52µs │    99.69ms
HTTP         │  100.0% │     599.69/s │     1.67ms │   608.20µs │     1.03ms │    98.64ms
WS           │  100.0% │     645.10/s │     1.55ms │   538.71µs │   797.57µs │    98.71ms 🏆

Performance Analysis:
  • WS is 8.2% faster than GRPC-UNIX in throughput
  • WS has the best P95 latency: 797.57µs

Legend: 🏆 Best Overall | ⚡ Highest Throughput | 🎯 Lowest Latency

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Test 5: Long Document
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │  Success │   Throughput │        Avg │        P50 │        P95 │        P99
─────────────┼──────────┼──────────────┼────────────┼────────────┼────────────┼──────────
GRPC         │  100.0% │     351.53/s │     2.84ms │     1.74ms │     2.56ms │    98.58ms
GRPC-UNIX    │  100.0% │     349.35/s │     2.86ms │     1.75ms │     2.75ms │    96.38ms
HTTP         │  100.0% │     372.97/s │     2.68ms │     1.65ms │     2.06ms │    99.42ms
WS           │  100.0% │     380.15/s │     2.63ms │     1.62ms │     2.07ms │    97.28ms 🏆

Performance Analysis:
  • WS is 8.8% faster than GRPC-UNIX in throughput
  • HTTP has the best P95 latency: 2.06ms

Legend: 🏆 Best Overall | ⚡ Highest Throughput | 🎯 Lowest Latency

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Test 6: News Headline
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │  Success │   Throughput │        Avg │        P50 │        P95 │        P99
─────────────┼──────────┼──────────────┼────────────┼────────────┼────────────┼──────────
GRPC         │  100.0% │    2816.31/s │   354.09µs │   374.68µs │   426.02µs │   506.93µs
GRPC-UNIX    │  100.0% │    2756.17/s │   361.54µs │   365.96µs │   422.44µs │   552.15µs
HTTP         │  100.0% │    3105.87/s │   321.06µs │   350.79µs │   456.05µs │   606.80µs
WS           │  100.0% │    3783.75/s │   263.55µs │   246.26µs │   377.36µs │   492.99µs 🏆

Performance Analysis:
  • WS is 37.3% faster than GRPC-UNIX in throughput
  • WS has the best P95 latency: 377.36µs

Legend: 🏆 Best Overall | ⚡ Highest Throughput | 🎯 Lowest Latency

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Test 7: Technical Article
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │  Success │   Throughput │        Avg │        P50 │        P95 │        P99
─────────────┼──────────┼──────────────┼────────────┼────────────┼────────────┼──────────
GRPC         │  100.0% │    1140.66/s │   875.57µs │   802.67µs │     1.25ms │     1.92ms
GRPC-UNIX    │  100.0% │    1178.18/s │   847.72µs │   786.66µs │     1.17ms │     1.30ms
HTTP         │  100.0% │    1182.78/s │   844.56µs │   777.15µs │     1.18ms │     1.33ms
WS           │  100.0% │    1356.38/s │   736.60µs │   710.86µs │   923.98µs │     1.08ms 🏆

Performance Analysis:
  • WS is 18.9% faster than GRPC in throughput
  • WS has the best P95 latency: 923.98µs

Legend: 🏆 Best Overall | ⚡ Highest Throughput | 🎯 Lowest Latency

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Test 8: Customer Support
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │  Success │   Throughput │        Avg │        P50 │        P95 │        P99
─────────────┼──────────┼──────────────┼────────────┼────────────┼────────────┼──────────
GRPC         │  100.0% │    1484.31/s │   672.37µs │   614.07µs │   881.21µs │     1.18ms
GRPC-UNIX    │  100.0% │    1732.21/s │   576.42µs │   543.52µs │   807.95µs │   950.28µs
HTTP         │  100.0% │    1615.21/s │   618.00µs │   547.41µs │   915.05µs │     1.02ms
WS           │  100.0% │    1996.43/s │   500.29µs │   482.22µs │   674.14µs │   894.17µs 🏆

Performance Analysis:
  • WS is 34.5% faster than GRPC in throughput
  • WS has the best P95 latency: 674.14µs

Legend: 🏆 Best Overall | ⚡ Highest Throughput | 🎯 Lowest Latency

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Test 9: Short Greeting
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │  Success │   Throughput │        Avg │        P50 │        P95 │        P99
─────────────┼──────────┼──────────────┼────────────┼────────────┼────────────┼──────────
GRPC         │  100.0% │    5750.81/s │   173.00µs │   183.06µs │   261.59µs │   277.08µs
GRPC-UNIX    │  100.0% │    7691.49/s │   129.42µs │   106.68µs │   232.04µs │   265.57µs
HTTP         │  100.0% │    7670.47/s │   129.67µs │   114.93µs │   224.46µs │   408.05µs
WS           │  100.0% │   12445.50/s │    79.98µs │    74.03µs │   119.06µs │   145.55µs 🏆

Performance Analysis:
  • WS is 116.4% faster than GRPC in throughput
  • WS has the best P95 latency: 119.06µs

Legend: 🏆 Best Overall | ⚡ Highest Throughput | 🎯 Lowest Latency

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Test 10: Product Description
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │  Success │   Throughput │        Avg │        P50 │        P95 │        P99
─────────────┼──────────┼──────────────┼────────────┼────────────┼────────────┼──────────
GRPC         │  100.0% │    2329.67/s │   428.38µs │   400.59µs │   591.65µs │   741.14µs
GRPC-UNIX    │  100.0% │    2233.62/s │   446.85µs │   416.86µs │   583.48µs │     1.09ms
HTTP         │  100.0% │    2412.39/s │   413.74µs │   365.76µs │   594.27µs │     1.41ms
WS           │  100.0% │    3013.37/s │   331.40µs │   311.80µs │   565.38µs │   656.79µs 🏆

Performance Analysis:
  • WS is 34.9% faster than GRPC-UNIX in throughput
  • WS has the best P95 latency: 565.38µs

Legend: 🏆 Best Overall | ⚡ Highest Throughput | 🎯 Lowest Latency

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Overall Protocol Performance Summary
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Protocol     │ Avg Throughput │  Avg Latency │       Wins
─────────────┼──────────────┼──────────────┼──────────
GRPC         │    1793.11/s │     989.94µs │ 0/10
GRPC-UNIX    │    1990.16/s │     986.04µs │ 0/10
HTTP         │    2061.02/s │     953.71µs │ 0/10
WS           │    2774.14/s │     867.12µs │ 10/10


✅ Benchmark completed successfully!