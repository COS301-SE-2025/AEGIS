
import React from "react";
import { ChromePicker, ColorResult } from "react-color";
import { useTheme } from "../../context/ThemeContext";

// Helper: Convert hex to HSL string (e.g., "210 100% 56%")
function hexToHSL(hex: string): string {
  // Remove # if present
  hex = hex.replace('#', '');
  let r = 0, g = 0, b = 0;
  if (hex.length === 3) {
    r = parseInt(hex[0] + hex[0], 16);
    g = parseInt(hex[1] + hex[1], 16);
    b = parseInt(hex[2] + hex[2], 16);
  } else if (hex.length === 6) {
    r = parseInt(hex.substring(0,2), 16);
    g = parseInt(hex.substring(2,4), 16);
    b = parseInt(hex.substring(4,6), 16);
  }
  r /= 255; g /= 255; b /= 255;
  const max = Math.max(r, g, b), min = Math.min(r, g, b);
  let h = 0, s = 0, l = (max + min) / 2;
  if (max !== min) {
    const d = max - min;
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
    switch(max){
      case r: h = (g - b) / d + (g < b ? 6 : 0); break;
      case g: h = (b - r) / d + 2; break;
      case b: h = (r - g) / d + 4; break;
    }
    h /= 6;
  }
  h = Math.round(h * 360);
  s = Math.round(s * 100);
  l = Math.round(l * 100);
  return `${h} ${s}% ${l}%`;
}

export const ColorPalettePicker = () => {
  const { palette, setPalette } = useTheme();

  // Helper to convert HSL string to hex for ChromePicker
  function hslToHex(hsl: string): string {
    // hsl: "210 100% 56%"
    const [h, s, l] = hsl.split(/[ %]+/).map(Number);
    const a = s * Math.min(l, 100 - l) / 10000;
    const f = (n: number) => {
      const k = (n + h / 30) % 12;
      const color = l - a * Math.max(Math.min(k - 3, 9 - k, 1), -1);
      return Math.round(255 * color / 100).toString(16).padStart(2, '0');
    };
    return `#${f(0)}${f(8)}${f(4)}`;
  }

  return (
    <div className="flex flex-wrap gap-8">
      <div>
        <p className="mb-2">Primary Color</p>
        <ChromePicker
          color={hslToHex(palette.primary)}
          onChange={(color: ColorResult) => setPalette({ primary: hexToHSL(color.hex) })}
        />
      </div>
      <div>
        <p className="mb-2">Secondary Color</p>
        <ChromePicker
          color={hslToHex(palette.secondary)}
          onChange={color => setPalette({ secondary: hexToHSL(color.hex) })}
        />
      </div>
      <div>
        <p className="mb-2">Background Color</p>
        <ChromePicker
          color={hslToHex(palette.background)}
          onChange={color => setPalette({ background: hexToHSL(color.hex) })}
        />
      </div>
      <div>
        <p className="mb-2">Foreground (Text) Color</p>
        <ChromePicker
          color={hslToHex(palette.foreground)}
          onChange={color => setPalette({ foreground: hexToHSL(color.hex) })}
        />
      </div>
      <div>
        <p className="mb-2">Card Color</p>
        <ChromePicker
          color={hslToHex(palette.card)}
          onChange={color => setPalette({ card: hexToHSL(color.hex) })}
        />
      </div>
    </div>
  );
};