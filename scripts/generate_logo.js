const fs = require('fs');
const path = require('path');
const { Resvg } = require('@resvg/resvg-js');

function buildLogoSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="1024" height="1024">
  <defs>
    <clipPath id="squircle-clip">
      <rect x="24" y="24" width="976" height="976" rx="220" />
    </clipPath>

    <!-- Optics & Telemetry Electric Cyan -->
    <linearGradient id="cyan-glow" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#00f5ff" />
      <stop offset="100%" stop-color="#0284c7" />
    </linearGradient>

    <!-- Emerald Qualification Indicator -->
    <linearGradient id="emerald-accent" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#34d399" />
      <stop offset="100%" stop-color="#059669" />
    </linearGradient>

    <!-- Raptor Beak Gold -->
    <linearGradient id="beak-gold" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#fbbf24" />
      <stop offset="100%" stop-color="#d97706" />
    </linearGradient>

    <!-- Titanium Plate -->
    <linearGradient id="titanium-plate" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#cbd5e1" />
      <stop offset="100%" stop-color="#64748b" />
    </linearGradient>

    <filter id="subtle-shadow" x="-10%" y="-10%" width="120%" height="120%">
      <feDropShadow dx="0" dy="16" stdDeviation="20" flood-color="#0f172a" flood-opacity="0.16" />
    </filter>
  </defs>

  <!-- 1. Luxury White Squircle Container -->
  <rect x="24" y="24" width="976" height="976" rx="220" fill="#ffffff" stroke="#e2e8f0" stroke-width="6" />

  <g clip-path="url(#squircle-clip)">
    <g transform="translate(512, 512)" filter="url(#subtle-shadow)">

      <!-- 2. Hexagonal Architectural Gateway Frame -->
      <polygon points="
        0,-390
        338,-195
        338,195
        0,390
        -338,195
        -338,-195
      " fill="none" stroke="#0f172a" stroke-width="40" stroke-linejoin="round" />

      <!-- Inner dashed telemetry guide -->
      <polygon points="
        0,-355
        307,-177
        307,177
        0,355
        -307,177
        -307,-177
      " fill="none" stroke="#00f5ff" stroke-width="4" opacity="0.4" stroke-dasharray="16, 12" />

      <!-- 3. Watertight Solid Base Silhouette (Zero-Leak Rule) -->
      <!-- Sharp Aerodynamic Peregrine Falcon Delta Wing Silhouette -->
      <path d="
        M 0 -315
        L 50 -265
        L 95 -295
        L 85 -225
        L 165 -165
        L 265 -55
        L 245 45
        L 275 95
        L 210 165
        L 135 240
        L 65 285
        L 0 340
        L -65 285
        L -135 240
        L -210 165
        L -275 95
        L -245 45
        L -265 -55
        L -165 -165
        L -85 -225
        L -95 -295
        L -50 -265
        Z
      " fill="#0b0f19" stroke="#0f172a" stroke-width="8" stroke-linejoin="round" />

      <!-- 4. Low-Poly Peregrine Falcon Origami Facets -->

      <!-- Crown & Crest -->
      <polygon points="0,-315 50,-265 0,-240" fill="#475569" />
      <polygon points="0,-315 -50,-265 0,-240" fill="#1e293b" />

      <polygon points="50,-265 95,-295 85,-225" fill="#64748b" />
      <polygon points="-50,-265 -95,-295 -85,-225" fill="#0f172a" />

      <polygon points="0,-240 50,-265 85,-225 35,-190 0,-195" fill="#334155" />
      <polygon points="0,-240 -50,-265 -85,-225 -35,-190 0,-195" fill="#1e293b" />

      <!-- Outer Ears & Temple Armor -->
      <polygon points="85,-225 165,-165 115,-130 35,-190" fill="#475569" />
      <polygon points="-85,-225 -165,-165 -115,-130 -35,-190" fill="#0f172a" />

      <!-- Supraorbital Brow Shelf (Sharp Predator Cant) -->
      <polygon points="0,-195 35,-190 85,-155 35,-145 0,-140" fill="#1e293b" />
      <polygon points="0,-195 -35,-190 -85,-155 -35,-145 0,-140" fill="#0b0f19" />

      <!-- Nasal Bridge -->
      <polygon points="0,-140 35,-145 22,-85 0,-80" fill="#475569" />
      <polygon points="0,-140 -35,-145 -22,-85 0,-80" fill="#1e293b" />

      <!-- Eye Sockets (Dark Obsidian Housing) -->
      <polygon points="35,-145 85,-155 105,-120 50,-120" fill="#082f49" />
      <polygon points="-35,-145 -85,-155 -105,-120 -50,-120" fill="#082f49" />

      <!-- Canted Predator Optics (Electric Cyan Almond Gaze) -->
      <polygon points="44,-140 78,-148 95,-124 55,-124" fill="url(#cyan-glow)" />
      <polygon points="-44,-140 -78,-148 -95,-124 -55,-124" fill="url(#cyan-glow)" />

      <!-- Slit Pupils -->
      <polygon points="66,-142 72,-134 70,-126 64,-134" fill="#021a36" />
      <polygon points="-66,-142 -72,-134 -70,-126 -64,-134" fill="#021a36" />

      <!-- Specular Highlight -->
      <circle cx="63" cy="-136" r="2" fill="#ffffff" />
      <circle cx="-63" cy="-136" r="2" fill="#ffffff" />

      <!-- Peregrine Malar Cheek Plates (Distinctive Falcon Facial Mark) -->
      <polygon points="85,-155 165,-165 145,-90 105,-120" fill="#0f172a" />
      <polygon points="-85,-155 -165,-165 -145,-90 -105,-120" fill="#0b0f19" />

      <polygon points="50,-120 105,-120 125,-55 55,-65 22,-85" fill="#1e293b" />
      <polygon points="-50,-120 -105,-120 -125,-55 -55,-65 -22,-85" fill="#0f172a" />

      <!-- Upper Beak Cere (Titanium) -->
      <polygon points="0,-80 22,-85 18,-45 0,-40" fill="url(#titanium-plate)" />
      <polygon points="0,-80 -22,-85 -18,-45 0,-40" fill="#475569" />

      <!-- Sharp Curved Raptor Beak Tip (Amber Falcon Hook) -->
      <polygon points="0,-40 18,-45 14,0 0,25" fill="url(#beak-gold)" />
      <polygon points="0,-40 -18,-45 -14,0 0,25" fill="#b45309" />

      <!-- Lower Beak Cutting Edge -->
      <polygon points="0,25 14,0 0,5" fill="#78350f" />
      <polygon points="0,25 -14,0 0,5" fill="#451a03" />

      <!-- Throat & Neck Armor -->
      <polygon points="0,5 18,-45 55,-65 45,35 0,55" fill="#1e293b" />
      <polygon points="0,5 -18,-45 -55,-65 -45,35 0,55" fill="#0f172a" />

      <polygon points="55,-65 125,-55 120,20 45,35" fill="#334155" />
      <polygon points="-55,-65 -125,-55 -120,20 -45,35" fill="#0f172a" />

      <!-- Swept Deltoids & Shoulders -->
      <polygon points="165,-165 265,-55 200,-35 145,-90" fill="#475569" />
      <polygon points="-165,-165 -265,-55 -200,-35 -145,-90" fill="#1e293b" />

      <!-- Outer Aerodynamic Wings (Wing Pinions) -->
      <polygon points="265,-55 245,45 190,15 200,-35" fill="#334155" />
      <polygon points="-265,-55 -245,45 -190,15 -200,-35" fill="#0f172a" />

      <polygon points="245,45 275,95 210,105 190,15" fill="#1e293b" />
      <polygon points="-245,45 -275,95 -210,105 -190,15" fill="#0b0f19" />

      <polygon points="145,-90 200,-35 190,15 120,20 125,-55" fill="#334155" />
      <polygon points="-145,-90 -200,-35 -190,15 -120,20 -125,-55" fill="#1e293b" />

      <!-- Central Breastplate: Supersonic Chevron Guidance Vector -->
      <polygon points="0,55 75,100 0,150" fill="url(#cyan-glow)" />
      <polygon points="0,55 -75,100 0,150" fill="#0284c7" />

      <polygon points="45,35 120,20 135,105 75,100" fill="#475569" />
      <polygon points="-45,35 -120,20 -135,105 -75,100" fill="#1e293b" />

      <polygon points="120,20 190,15 210,105 135,105" fill="#334155" />
      <polygon points="-120,20 -190,15 -210,105 -135,105" fill="#0f172a" />

      <!-- Secondary Qualification Emerald Chevron -->
      <polygon points="0,155 50,190 0,225" fill="url(#emerald-accent)" />
      <polygon points="0,155 -50,190 0,225" fill="#059669" />

      <polygon points="75,100 135,105 110,190 50,190 0,150" fill="#1e293b" />
      <polygon points="-75,100 -135,105 -110,190 -50,190 -0,150" fill="#0f172a" />

      <!-- Lower Wing Blades -->
      <polygon points="210,105 275,95 210,165 155,160 135,105" fill="#1e293b" />
      <polygon points="-210,105 -275,95 -210,165 -155,160 -135,105" fill="#0b0f19" />

      <polygon points="135,105 155,160 110,190" fill="#334155" />
      <polygon points="-135,105 -155,160 -110,190" fill="#0f172a" />

      <polygon points="210,165 135,240 110,190 155,160" fill="#1e293b" />
      <polygon points="-210,165 -135,240 -110,190 -155,160" fill="#0b0f19" />

      <!-- Tail Stabilizer & Vector Rudder -->
      <polygon points="0,230 45,260 0,295" fill="#334155" />
      <polygon points="0,230 -45,260 0,295" fill="#1e293b" />

      <polygon points="50,190 110,190 85,255 45,260" fill="#475569" />
      <polygon points="-50,190 -110,190 -85,255 -45,260" fill="#0f172a" />

      <polygon points="110,190 135,240 65,285 85,255" fill="#1e293b" />
      <polygon points="-110,190 -135,240 -65,285 -85,255" fill="#0b0f19" />

      <!-- Rudder Apex / Trailing Edge -->
      <polygon points="0,295 65,285 0,340" fill="#1e293b" />
      <polygon points="0,295 -65,285 0,340" fill="#0b0f19" />

      <polygon points="45,260 85,255 65,285 0,295" fill="#334155" />
      <polygon points="-45,260 -85,255 -65,285 0,295" fill="#1e293b" />

      <!-- HUD Telemetry Nodes -->
      <circle cx="0" cy="150" r="4" fill="#00f5ff" />
      <circle cx="75" cy="100" r="3" fill="#00f5ff" opacity="0.85" />
      <circle cx="-75" cy="100" r="3" fill="#00f5ff" opacity="0.85" />
      <circle cx="0" cy="225" r="3.5" fill="#34d399" />

    </g>
  </g>
</svg>`;
}

async function renderLogo(outputDir) {
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }

  const svg = buildLogoSvg();
  const svgPath = path.join(outputDir, 'logo.svg');
  const pngPath = path.join(outputDir, 'logo.png');

  fs.writeFileSync(svgPath, svg, 'utf8');

  const resvg = new Resvg(svg, { fitTo: { mode: 'width', value: 1024 } });
  const pngData = resvg.render().asPng();
  fs.writeFileSync(pngPath, pngData);

  console.log(`✓ Successfully rendered Peregrine Falcon logo into ${outputDir}`);
}

renderLogo(path.join(__dirname, '..', 'docs', 'images')).catch((err) => {
  console.error('Error rendering logo:', err);
  process.exit(1);
});
