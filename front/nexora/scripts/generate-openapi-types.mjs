import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import openapiTS, { astToString } from "openapi-typescript";
import jsYaml from "js-yaml";
import swagger2openapi from "swagger2openapi";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const projectRoot = path.resolve(__dirname, "..");

const schemaSource = process.env.OPENAPI_SCHEMA || "../../swagger.yaml";
const outputFile = process.env.OPENAPI_TYPES_OUTPUT || "./src/api/generated/swagger-types.ts";

const resolvedSchemaSource = schemaSource.startsWith("http://") || schemaSource.startsWith("https://")
    ? schemaSource
    : path.resolve(projectRoot, schemaSource);
const resolvedOutputFile = path.resolve(projectRoot, outputFile);

const rawSchema =
    resolvedSchemaSource.startsWith("http://") || resolvedSchemaSource.startsWith("https://")
        ? null
        : jsYaml.load(await readFile(resolvedSchemaSource, "utf8"));

const schemaInput =
    resolvedSchemaSource.startsWith("http://") || resolvedSchemaSource.startsWith("https://")
        ? resolvedSchemaSource
        : rawSchema?.swagger === "2.0"
            ? (await swagger2openapi.convertObj(rawSchema, { patch: true })).openapi
            : rawSchema;

const ast = await openapiTS(schemaInput);
const contents = astToString(ast);

await mkdir(path.dirname(resolvedOutputFile), { recursive: true });
await writeFile(resolvedOutputFile, contents, "utf8");

console.log(`Generated OpenAPI types from ${resolvedSchemaSource}`);
console.log(`Output: ${resolvedOutputFile}`);

