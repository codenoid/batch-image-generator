<!-- App.svelte -->
<script>
import {
	onMount,
	afterUpdate
} from 'svelte';
import logo from './assets/images/logo-universal.png'
import {
	Greet,
	Proceed
} from '../wailsjs/go/main/App.js'
import {
	TextBox,
	TextBoxButton,
	Button,
	ProgressRing
} from "fluent-svelte";
import { 
	ComboBox,
	TextInput
} from "carbon-components-svelte";
import "fluent-svelte/theme.css";
import "carbon-components-svelte/css/g80.css";

let resultText = "Please enter your name below 👇"
let name

function greet() {
	Greet(name).then(result => resultText = result)
}

let canvas;
let context;
let img = new Image();
let rectangles = [];
let selectedRect = -1; // initially, no rectangle is selected
let drag = false;
let rectId = 0; // counter for rectangle id
let csvHeaders = [];
let csvDelimiter = ","
let csvContent = '';
let templateBase64 = '';
let customFont;
let customFontFile = {};

let fonts = ["Arial"];

let imageFileName = '';
let csvFileName = '';

function loadImage(e) {
	const reader = new FileReader();
	reader.onload = (e) => {
		templateBase64 = reader.result.toString();
		img.onload = function() {
			canvas.width = this.width;
			canvas.height = this.height;
			updateCanvas();
		};
		img.src = e.target.result;
	};
	imageFileName = e.target.files[0].name;
	reader.readAsDataURL(e.target.files[0]);
}

function loadCSV(e) {
    const reader = new FileReader();
    reader.onload = (e) => {
        csvHeaders = e.target.result.split('\n')[0].split(csvDelimiter);
        csvContent = e.target.result;  // Store the content
    };
    csvFileName = e.target.files[0].name;
    reader.readAsText(e.target.files[0]);
}

function loadFont(e) {
    const reader = new FileReader();
    customFont = e.target.files[0].name.split('.')[0]; // Get the font name from the file name
	fonts = [customFont, ...fonts];
    reader.onload = function(event) {
        const font = new FontFace(customFont, event.target.result);
		customFontFile[customFont] = event.target.result;
        font.load().then(function(loadedFace) {
            document.fonts.add(loadedFace);
			rectangles[selectedRect].font = customFont;
    		updateCanvas();
        });
    };
    reader.readAsArrayBuffer(e.target.files[0]);

	const reader2 = new FileReader();
	reader2.onload = (e) => {
		customFontFile[customFont] = e.target.result.toString().split(',')[1];;
	};
	reader2.readAsDataURL(e.target.files[0]);
}

onMount(() => {
	context = canvas.getContext('2d');
});

function updateCanvas() {
	context.clearRect(0, 0, canvas.width, canvas.height);
	context.drawImage(img, 0, 0, canvas.width, canvas.height);
	rectangles.forEach((rect, index) => {
		context.beginPath();
		context.rect(rect.startX, rect.startY, rect.w, rect.h);

		// Set the stroke color based on whether the rectangle is selected
		context.strokeStyle = index === selectedRect ? '#51ff0d' : 'black';
		context.stroke();

		context.fillStyle = rect.color || 'black';
		context.font = `${rect.fontSize}px ${rect.font || 'Arial'}`;
		let rectText = rect.csv_key;
		if (rect.transform != "") {
			rectText += ` (${rect.transform})`
		}
		let textX;
		switch (rect.textAlign) {
			case 'center':
				textX = rect.startX + rect.w / 2;
				break;
			case 'right':
				textX = rect.startX + rect.w;
				break;
			default:
				textX = rect.startX;
		}
		context.textAlign = rect.textAlign || 'left';
		context.fillText(rectText, textX, rect.startY + rect.h);
	});
}

function createNewRectangle() {
	rectangles = [...rectangles, {
		id: rectId++,
		startX: 0,
		startY: 0,
		w: 0,
		h: 0,
		csv_key: 'Select CSV Key',
		color: '#000000',
		font: 'Arial',
		fontContent: '',
		textAlign: 'left',
		fontSize: 16,
		transform: ''
	}];
	selectedRect = rectangles.length - 1;
}

function selectRectangle(index) {
	selectedRect = index;
}

function deleteRectangle(index) {
	rectangles = rectangles.filter((_, i) => i !== index);
	selectedRect = -1; // deselect the rectangle
	updateCanvas();
}

function mouseDown(e) {
    const bounds = canvas.getBoundingClientRect();
    const x = e.clientX - bounds.left;
    const y = e.clientY - bounds.top;

    // Check if a rectangle was clicked and set it as selected
    for (let i = 0; i < rectangles.length; i++) {
        let rect = rectangles[i];
        if (
            x >= rect.startX &&
            x <= rect.startX + rect.w &&
            y >= rect.startY &&
            y <= rect.startY + rect.h
        ) {
            // Rectangle was clicked, select it
            selectedRect = i;
            drag = false;
			updateCanvas();
            return;
        }
    }

    // If no rectangle was clicked but a rectangle is selected, start dragging
    if (selectedRect >= 0 && selectedRect < rectangles.length) {
        rectangles[selectedRect].startX = x;
        rectangles[selectedRect].startY = y;
        drag = true;
    }
}

function mouseUp() {
	drag = false;
}

function mouseMove(e) {
	if (drag && selectedRect >= 0 && selectedRect < rectangles.length) {
		const bounds = canvas.getBoundingClientRect();
		rectangles[selectedRect].w = (e.clientX - bounds.left) - rectangles[selectedRect].startX;
		rectangles[selectedRect].h = (e.clientY - bounds.top) - rectangles[selectedRect].startY;
		// console.log(`Rectangle location: x=${rectangles[selectedRect].startX}, y=${rectangles[selectedRect].startY}, width=${rectangles[selectedRect].w}, height=${rectangles[selectedRect].h}`);
		updateCanvas();
	}
}

let loading = false;
async function proceed() {
    // Convert canvas to base64-encoded image
    let image = templateBase64.split("base64,")[1];

    // CSV file content as string is now directly accessible
    let csv = csvContent;

    // JSON stringified rectangles
	for (let item of rectangles) {
		item.fontContent = customFontFile[item.font];
	}
    let placeholder = JSON.stringify(rectangles);

    let output = {
        image,
        csv,
        placeholder
    };

	loading = true;
	try {
		await Proceed(image, placeholder, csv);
		loading = false;
		alert("Done!");
	} catch (e) {
		alert("Error:" + e.message);
	}
	loading = false;
    return output;
}

function reset() {
	// Here, you reset all your variables to their initial states
	resultText = "Please enter your name below 👇";
	name = '';
	rectangles = [];
	selectedRect = -1;
	rectId = 0;
	csvHeaders = [];
	csvDelimiter = ",";
	csvContent = '';
	templateBase64 = '';
	customFont = '';
	customFontFile = {};
	fonts = ["Arial"];
	imageFileName = '';
	csvFileName = '';
	loading = false;
	// Reload the canvas
	context.clearRect(0, 0, canvas.width, canvas.height);
	if (img.src) {
		context.drawImage(img, 0, 0, canvas.width, canvas.height);
	}
}

</script>

<div class="row">
	<div class="five columns">
		<div class="row">
			<div class="six columns">
				<div style="display: none;">
					<input type="file" on:change="{loadImage}" id="image-template" accept="image/*">
					<input type="file" on:change="{loadCSV}" id="image-data" accept=".csv">
				</div>
				<label>Select Image Template</label>
				<TextBox bind:value="{imageFileName}" placeholder="Select image template jpg/png/webp/etc." on:click={() => document.getElementById('image-template').click()}>
					<TextBoxButton slot="buttons" on:click={() => document.getElementById('image-template').click()}>
						<svg fill="currentColor" aria-hidden="true" width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M5.25 3.5h13.5a.75.75 0 0 0 .1-1.5H5.25a.75.75 0 0 0-.1 1.49h.1ZM11.88 22H12a1 1 0 0 0 1-.88V8.4l3.3 3.3a1 1 0 0 0 1.31.08l.1-.09a1 1 0 0 0 .08-1.32l-.08-.1-5-4.99a1 1 0 0 0-1.32-.08l-.1.08-5 5a1 1 0 0 0 1.32 1.5l.1-.09L11 8.42V21a1 1 0 0 0 .88 1Z" fill="currentColor"></path></svg>
					</TextBoxButton>
				</TextBox>
				<br>
				<label>Select CSV to Load</label>
				<TextBox  bind:value="{csvFileName}" placeholder="Select csv file with header...." on:click={() => document.getElementById('image-data').click()}>
					<TextBoxButton slot="buttons">
						<svg fill="currentColor" aria-hidden="true" width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M5.25 3.5h13.5a.75.75 0 0 0 .1-1.5H5.25a.75.75 0 0 0-.1 1.49h.1ZM11.88 22H12a1 1 0 0 0 1-.88V8.4l3.3 3.3a1 1 0 0 0 1.31.08l.1-.09a1 1 0 0 0 .08-1.32l-.08-.1-5-4.99a1 1 0 0 0-1.32-.08l-.1.08-5 5a1 1 0 0 0 1.32 1.5l.1-.09L11 8.42V21a1 1 0 0 0 .88 1Z" fill="currentColor"></path></svg>
					</TextBoxButton>
				</TextBox>
				<br>
				<label for="">CSV Delimiter</label>
				<TextBox style="display:inline-block;width:20px;" bind:value={csvDelimiter} />
				<br>
				<Button variant="accent" style="width:100%;" disabled={loading} on:click={proceed}>
					{#if loading}<ProgressRing size={20}/>&nbsp;{/if}
					{loading ? 'Processing...' : 'Proceed'}
				</Button>
				<br><br>
				<Button variant="standard" style="width:100%;" on:click={reset}>
					Reset
				</Button>				  
			</div>
			<div class="six columns">
				<label for="">Create New Placeholder</label>
				<Button variant="accent" style="width:100%;" on:click="{createNewRectangle}">New Placeholder</Button>
				<br><br>
				{#if rectangles.length > 0 && selectedRect >= 0}
				<label>CSV Key</label>
				<select style="width:100%;" bind:value="{rectangles[selectedRect].csv_key}" on:change="{updateCanvas}">
					<option value="Select CSV Key" selected>Select CSV Key</option>
					{#each csvHeaders as header}
					<option value="{header}">{header}</option>
					{/each}
				</select><br>
				<br>
				<label>Font:</label><br>
				<input style="display:none;" type="file" on:change="{loadFont}" id="font-data" accept=".ttf">
				<TextBox placeholder="Upload custom .ttf font." on:click={() => document.getElementById('font-data').click()}>
					<TextBoxButton slot="buttons" on:click={() => document.getElementById('font-data').click()}>
						<svg fill="currentColor" aria-hidden="true" width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M5.25 3.5h13.5a.75.75 0 0 0 .1-1.5H5.25a.75.75 0 0 0-.1 1.49h.1ZM11.88 22H12a1 1 0 0 0 1-.88V8.4l3.3 3.3a1 1 0 0 0 1.31.08l.1-.09a1 1 0 0 0 .08-1.32l-.08-.1-5-4.99a1 1 0 0 0-1.32-.08l-.1.08-5 5a1 1 0 0 0 1.32 1.5l.1-.09L11 8.42V21a1 1 0 0 0 .88 1Z" fill="currentColor"></path></svg>
					</TextBoxButton>
				</TextBox>
				<br>
				<select style="width:100%" bind:value="{rectangles[selectedRect].font}" on:change="{updateCanvas}">
					{#each fonts as font}
					<option value="{font}">{font}</option>
					{/each}
				</select>
				<br><br>
				<div class="row">
					<div class="six columns">
						<label>
							Font Size:
							<TextBox type="number" min="1" bind:value="{rectangles[selectedRect].fontSize}" on:change="{updateCanvas}"/>
						</label>
					</div>
					<div class="six columns">
						<label>
							Color: <br>
							<input type="color" bind:value="{rectangles[selectedRect].color}" on:change="{updateCanvas}">
						</label>
					</div>
				</div>
				<br><br><br><br>
				<div class="row" style="display: block !important;">
					<div class="six columns">
						<label>
							Text Align:
							<select style="width:100%;" bind:value="{rectangles[selectedRect].textAlign}" on:change="{updateCanvas}">
								<option value="left">Left</option>
								<option value="center">Center</option>
								<option value="right">Right</option>
							</select>
						</label>
					</div>
					<div class="six columns">
						<label>
							Transform:
							<select style="width:100%;" placeholder="Select transformation" bind:value="{rectangles[selectedRect].transform}" on:change="{updateCanvas}">
								<option value="">None</option>
								<option value="uppercase">Uppercase</option>
								<option value="qrcode">QR-Code</option>
								<option value="avatar">Avatar</option>
								<option value="initial-avatar">Initial Avatar</option>
							</select>
						</label>
					</div>
				</div>
				<Button style="width:100%;margin-top:10px;" on:click="{() => deleteRectangle(rectangles[selectedRect].id)}">Delete</Button>
				<!-- Rectangle {rect.id}: x={rect.startX}, y={rect.startY}, width={rect.w}, height={rect.h} -->
				{/if}
			</div>
		</div>
	</div>
	<div class="seven columns">
		<br><br><br>
		<div style="max-height:600px;overflow-y:auto;">
			<canvas
				bind:this="{canvas}"
				on:mousedown="{mouseDown}"
				on:mouseup="{mouseUp}"
				on:mousemove="{mouseMove}"
				width="800"
				style="border:1px solid black;">
			</canvas>
		</div>
	</div>
</div>
