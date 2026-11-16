const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const puppeteer = require('puppeteer-core');

const EDGE_PATH = 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe';
const APP_URL = 'http://localhost:9393';
const OUTPUT_DIR = path.join(__dirname, '..', 'docs', 'images');
const FRAMES_DIR = path.join(__dirname, '..', 'docs', 'images', 'temp_frames');

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function capture() {
  if (!fs.existsSync(OUTPUT_DIR)) fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  if (!fs.existsSync(FRAMES_DIR)) fs.mkdirSync(FRAMES_DIR, { recursive: true });

  console.log('Launching browser with Edge executable...');
  const browser = await puppeteer.launch({
    executablePath: EDGE_PATH,
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--window-size=1600,1000'],
  });

  try {
    const page = await browser.newPage();
    await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 2 });

    console.log(`Navigating to ${APP_URL}...`);
    await page.goto(APP_URL, { waitUntil: 'networkidle0' });
    await sleep(1500);

    // 1. Desktop Dashboard Full View
    console.log('Capturing Desktop Dashboard (docs/images/cockpit_dashboard.png)...');
    await page.screenshot({
      path: path.join(OUTPUT_DIR, 'cockpit_dashboard.png'),
      fullPage: false,
    });

    // 2. Lead Detail & Latency Waterfall Focus
    console.log('Capturing Lead Detail & Latency Waterfall (docs/images/lead_detail_waterfall.png)...');
    // Click on the Apex Clearing lead (hot enterprise lead)
    const leadItems = await page.$$('div[class*="cursor-pointer"]');
    if (leadItems.length > 0) {
      await leadItems[0].click();
      await sleep(1000);
    }
    await page.screenshot({
      path: path.join(OUTPUT_DIR, 'lead_detail_waterfall.png'),
      fullPage: false,
    });

    // 3. Inbound Simulator Modal View
    console.log('Capturing Inbound Simulator (docs/images/simulator_settings.png)...');
    // Click on the Simulate Lead button
    const simulateButton = await page.$('button[class*="bg-emerald-600"]');
    if (simulateButton) {
      await simulateButton.click();
      await sleep(1000);
      await page.screenshot({
        path: path.join(OUTPUT_DIR, 'simulator_settings.png'),
        fullPage: false,
      });

      // Close modal
      const closeButtons = await page.$$('button');
      for (const btn of closeButtons) {
        const text = await page.evaluate(el => el.textContent, btn);
        if (text && text.includes('Cancel')) {
          await btn.click();
          break;
        }
      }
      await sleep(500);
    }

    // 4. Mobile Responsive Layout View
    console.log('Capturing Mobile View (docs/images/mobile_cockpit.png)...');
    await page.setViewport({ width: 414, height: 896, deviceScaleFactor: 2 });
    await sleep(800);
    await page.screenshot({
      path: path.join(OUTPUT_DIR, 'mobile_cockpit.png'),
      fullPage: false,
    });

    // 5. Generate Animated Demo GIF Frames
    console.log('Recording interactive frames for formpilot_demo.gif...');
    await page.setViewport({ width: 1280, height: 760, deviceScaleFactor: 1.5 });
    await page.goto(APP_URL, { waitUntil: 'networkidle0' });
    await sleep(1000);

    let frameCount = 0;
    async function snapFrame(repeat = 1) {
      for (let i = 0; i < repeat; i++) {
        const framePath = path.join(FRAMES_DIR, `frame_${String(frameCount++).padStart(4, '0')}.png`);
        await page.screenshot({ path: framePath });
      }
    }

    // Initial resting dashboard
    await snapFrame(10);

    // Open simulator modal
    const simBtn = await page.$('button[class*="bg-emerald-600"]');
    if (simBtn) {
      await simBtn.click();
      await sleep(400);
      await snapFrame(8);

      // Click on FinTech preset
      const presetButtons = await page.$$('button');
      for (const btn of presetButtons) {
        const text = await page.evaluate(el => el.textContent, btn);
        if (text && text.includes('FinTech Operations RFP')) {
          await btn.click();
          break;
        }
      }
      await sleep(800);
      await snapFrame(12);

      // Close or inspect modal result
      const closeBtns = await page.$$('button');
      for (const btn of closeBtns) {
        const text = await page.evaluate(el => el.textContent, btn);
        if (text && (text.includes('Inspect in Cockpit') || text.includes('Cancel'))) {
          await btn.click();
          break;
        }
      }
      await sleep(600);
      await snapFrame(10);
    }

    // Select different leads to show real-time waterfall latency and tone switching
    const currentLeads = await page.$$('div[class*="cursor-pointer"]');
    if (currentLeads.length > 1) {
      // Click second lead
      await currentLeads[1].click();
      await sleep(500);
      await snapFrame(8);

      // Click first lead
      await currentLeads[0].click();
      await sleep(500);
      await snapFrame(8);
    }

    // Switch outreach draft tone
    const toneButtons = await page.$$('button');
    for (const btn of toneButtons) {
      const text = await page.evaluate(el => el.textContent, btn);
      if (text && text.includes('Consultative')) {
        await btn.click();
        await sleep(500);
        await snapFrame(10);
        break;
      }
    }

    // Final hold
    await snapFrame(10);

    console.log(`Captured ${frameCount} frames in temp_frames. Generating GIF with ffmpeg...`);
    const gifOutput = path.join(OUTPUT_DIR, 'formpilot_demo.gif');
    const palettePath = path.join(FRAMES_DIR, 'palette.png');

    // Generate high quality palette and GIF with ffmpeg
    execSync(`ffmpeg -y -framerate 6 -i "${FRAMES_DIR}\\frame_%04d.png" -vf "fps=6,scale=960:-1:flags=lanczos,palettegen=stats_mode=diff" "${palettePath}"`, { stdio: 'inherit' });
    execSync(`ffmpeg -y -framerate 6 -i "${FRAMES_DIR}\\frame_%04d.png" -i "${palettePath}" -lavfi "fps=6,scale=960:-1:flags=lanczos [x]; [x][1:v] paletteuse=dither=bayer:bayer_scale=3" "${gifOutput}"`, { stdio: 'inherit' });

    console.log(`✓ Successfully generated formpilot_demo.gif at ${gifOutput}`);

    // Cleanup temp frames
    fs.rmSync(FRAMES_DIR, { recursive: true, force: true });
    console.log('✓ Cleaned up temporary frames.');
  } finally {
    await browser.close();
  }
}

capture().catch((err) => {
  console.error('Asset capture failed:', err);
  process.exit(1);
});
