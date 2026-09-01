export const humanAppearanceColors = {
  skin: {
    porcelain: "#f2d0b5",
    fair: "#e2b38f",
    tan: "#c78b63",
    brown: "#966044",
    deep: "#65402f",
  },
  hair: {
    black: "#211b18",
    darkBrown: "#3b2921",
    chestnut: "#6a412d",
    auburn: "#7b3d2b",
    golden: "#c39b58",
    gray: "#85817c",
  },
  tunic: {
    slateBlue: "#4f7596",
    forest: "#4f7048",
    rust: "#a4573e",
    mustard: "#b28a42",
    plum: "#71566f",
    teal: "#3f7773",
    burgundy: "#7a3f4b",
  },
  trousers: {
    charcoal: "#343a43",
    navy: "#34435a",
    umber: "#59483b",
    olive: "#505644",
    taupe: "#675d54",
  },
  shoes: {
    darkBrown: "#49352b",
    oxblood: "#553230",
    charcoal: "#303238",
    tan: "#72533c",
  },
} as const;

export type SkinTone = keyof typeof humanAppearanceColors.skin;
export type HairStyle = "cropped" | "swept" | "bob" | "curls";
export type HairColor = keyof typeof humanAppearanceColors.hair;
export type TunicColor = keyof typeof humanAppearanceColors.tunic;
export type TrousersColor = keyof typeof humanAppearanceColors.trousers;
export type ShoeColor = keyof typeof humanAppearanceColors.shoes;

export type HumanAppearance = {
  skinTone: SkinTone;
  hairStyle: HairStyle;
  hairColor: HairColor;
  tunicColor: TunicColor;
  trousersColor: TrousersColor;
  shoeColor: ShoeColor;
};
