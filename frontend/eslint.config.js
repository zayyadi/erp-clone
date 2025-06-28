import globals from "globals";
import tseslint from "typescript-eslint";
import pluginReact from "eslint-plugin-react";
import pluginPrettier from "eslint-plugin-prettier";
import configPrettier from "eslint-config-prettier"; // Renamed to avoid conflict

export default [
  {
    // Global ignores
    ignores: ["build/", ".vscode/", "node_modules/"],
  },
  ...tseslint.configs.recommended, // Apply TypeScript recommended rules
  {
    // Base configuration for all JavaScript/TypeScript files
    files: ["src/**/*.{js,jsx,ts,tsx}"],
    languageOptions: {
      parserOptions: {
        ecmaFeatures: { jsx: true },
        ecmaVersion: "latest",
        sourceType: "module",
        project: "./tsconfig.json", // Needed for some TypeScript rules
      },
      globals: {
        ...globals.browser,
        ...globals.es2021,
        // ...globals.node, // Uncomment if you use Node.js specific globals
      },
    },
    plugins: {
      react: pluginReact,
      prettier: pluginPrettier,
      // "@typescript-eslint" plugin is often included via tseslint.configs.recommended
    },
    rules: {
      ...configPrettier.rules, // Disables ESLint rules that conflict with Prettier
      "prettier/prettier": "warn", // Show Prettier issues as warnings

      // React specific rules
      "react/jsx-uses-react": "off",
      "react/react-in-jsx-scope": "off",
      "react/prop-types": "off", // Using TypeScript

      // TypeScript specific rules
      "@typescript-eslint/no-unused-vars": [
        "warn",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_" },
      ],
      "@typescript-eslint/explicit-function-return-type": "off",
      "@typescript-eslint/no-explicit-any": "warn",

      // General good practices
      "no-console": "warn",
      "no-debugger": "warn",
    },
    settings: {
      react: {
        version: "detect",
      },
    },
  },
  {
    // Configuration for Jest tests
    files: ["src/**/*.test.{js,jsx,ts,tsx}"],
    languageOptions: {
      globals: {
        ...globals.jest,
      },
    },
    // rules: { // Add Jest specific rules here if needed
    // }
  }
];
