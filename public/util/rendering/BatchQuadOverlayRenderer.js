"use strict";

import defaultTo from "lodash/defaultTo";
import VertexBuffer from "lumo/src/webgl/vertex/VertexBuffer";
import WebGLOverlayRenderer from "lumo/src/renderer/overlay/WebGLOverlayRenderer";

// Constants

/**
 * Shader GLSL source.
 * //normal rendering program
 * @private
 * @constant {object}
 */
const SHADER_GLSL = {
  vert: `
		precision highp float;
    attribute vec2 aPosition;
    attribute vec4 aColor;
		uniform vec2 uViewOffset;
		uniform float uScale;
    uniform mat4 uProjectionMatrix;
    varying vec4 oColor;
		void main() {
			vec2 wPosition = (aPosition * uScale) - uViewOffset;
      gl_Position = uProjectionMatrix * vec4(wPosition, 0.0, 1.0);
      oColor = aColor;
		}
		`,
  frag: `
    precision highp float;
    varying vec4 oColor;
		void main() {
      gl_FragColor = oColor;
      gl_FragColor.rgb *= gl_FragColor.a; // premultiplied alpha
		}
		`,
};
// picking shader, used to render the quad's id to screen
const PICKING_SHADER = {
  vert: `
  precision highp float;
  attribute vec2 aPosition;
  attribute vec4 aColor;
  attribute vec4 id;
  uniform vec2 uViewOffset;
  uniform float uScale;
  uniform mat4 uProjectionMatrix;
  varying vec4 oId;
  void main() {
    vec2 wPosition = (aPosition * uScale) - uViewOffset;
    gl_Position = uProjectionMatrix * vec4(wPosition, 0.0, 1.0);
    oId = id;
  }`,
  frag: `
    precision highp float;
    varying vec4 oId;
		void main() {
			gl_FragColor = oId;
		}
		`,
};

// create inline float array of all the vertex data: position, color, id
const getVertexArray = function (points) {
  const numOfAttrs = 10;
  const vertices = new Float32Array(points.length * numOfAttrs);
  for (let i = 0; i < points.length; i++) {
    vertices[i * numOfAttrs] = points[i].x;
    vertices[i * numOfAttrs + 1] = points[i].y;
    vertices[i * numOfAttrs + 2] = points[i].r;
    vertices[i * numOfAttrs + 3] = points[i].g;
    vertices[i * numOfAttrs + 4] = points[i].b;
    vertices[i * numOfAttrs + 5] = points[i].a;
    vertices[i * numOfAttrs + 6] = points[i].iR;
    vertices[i * numOfAttrs + 7] = points[i].iG;
    vertices[i * numOfAttrs + 8] = points[i].iB;
    vertices[i * numOfAttrs + 9] = points[i].iA;
  }
  return vertices;
};
// creates the gl buffers and creates the attrib pointers
const createBuffers = function (overlay, points) {
  const vertices = getVertexArray(points);
  const floatByteSize = 4;
  const vertSize = 2; // x,y
  const colorSize = 4;
  const vertexBuffer = new VertexBuffer(
    overlay.gl,
    vertices,
    {
      0: {
        size: 2,
        type: "FLOAT",
        byteOffset: 0,
      }, // vertex pointer
      1: {
        size: 4,
        type: "FLOAT",
        byteOffset: vertSize * floatByteSize,
      }, // color pointer
      2: {
        size: 4,
        type: "FLOAT",
        byteOffset: (colorSize + vertSize) * floatByteSize,
      }, // id pointer
    },
    {
      mode: "TRIANGLES",
      count: vertices.length / 10, // number of vertices to draw vertices has x,y therefore /2
    }
  );

  return {
    vertex: vertexBuffer,
  };
};
export const EVENT_TYPES = {
  MOUSE_HOVER: "mousehover",
  MOUSE_CLICK: "mouseclick",
};
/**
 * Class representing a Batchquad Renderer.
 */
export class BatchQuadOverlayRenderer extends WebGLOverlayRenderer {
  /**
   * Instantiates a new quadOverlayRenderer object.
   *
   * @param {object} options - The overlay options.
   * @param {Array} options.quadColor - The color of the line.
   */
  constructor(options = {}) {
    super(options);
    this.quadColor = defaultTo(options.quadColor, [1.0, 0.4, 0.1, 0.8]);
    this.shader = null;
    this.quads = null;
    this.pickingShader = null;
    //framebuffer
    this.targetTexture = null;
    this.depthBuffer = null;
    this.fbo = null;
    this.fboDimensions = { width: 0, height: 0 };
    this.callbacks = { mousehover: [], mouseclick: [] };
    const secondsToMillis = 1000;
    this.hoverThreshold = defaultTo(
      options.hoverThreshold,
      1 * secondsToMillis
    ); // two seconds hover threshold
    this.BACKGROUND_ID = -1;
    this.hoverTimeoutId = null;
    this.boundOnMove = this.onMove.bind(this);
    this.boundOnClick = this.onClick.bind(this);
  }
  /**
   * Executed when the overlay is attached to a plot.
   *
   * @param {Plot} plot - The plot to attach the overlay to.
   *
   * @returns {BatchquadOverlayRenderer} The overlay object, for chaining.
   */
  onAdd(plot) {
    super.onAdd(plot);
    this.shader = this.createShader(SHADER_GLSL);
    this.pickingShader = this.createShader(PICKING_SHADER);
    this.enableInteractions();
    this.gl.canvas,
      addEventListener("mouseleave", () => {
        clearTimeout(this.hoverTimeoutId);
      }); // if mouse leaves canvas remove hover timer so it doesn't trigger
    this.createFBO();
    return this;
  }

  /**
   * Executed when the overlay is removed from a plot.
   *
   * @param {Plot} plot - The plot to remove the overlay from.
   *
   * @returns {BatchquadOverlayRenderer} The overlay object, for chaining.
   */
  onRemove(plot) {
    super.onRemove(plot);
    this.shader = null;
    this.pickingShader = null;
    return this;
  }
  /**
   * disables quad interactions such as click and hover
   */
  disableInteractions() {
    clearTimeout(this.hoverTimeoutId); // cleanup
    this.gl.canvas.removeEventListener("mouseup", this.boundOnClick);
    this.gl.canvas.removeEventListener("mousemove", this.boundOnMove);
  }
  /**
   * enables quad interactions such as click and hover
   */
  enableInteractions() {
    this.gl.canvas.addEventListener("mouseup", this.boundOnClick);
    this.gl.canvas.addEventListener("mousemove", this.boundOnMove);
  }
  /**
   * Generate any underlying buffers.
   *
   * @returns {BatchquadOverlayRenderer} The overlay object, for chaining.
   */
  refreshBuffers() {
    const clipped = this.overlay.getClippedGeometry();
    if (clipped) {
      this.quads = clipped.map((points) => {
        // generate the buffer
        return createBuffers(this, points);
      });
    } else {
      this.quads = null;
    }
  }
  // normal render for human viewing
  renderColor() {
    const gl = this.gl;
    const shader = this.shader;
    const quads = this.quads;
    const plot = this.overlay.plot;
    const cell = plot.cell;
    const proj = this.getOrthoMatrix();
    const scale = Math.pow(2, plot.zoom - cell.zoom);
    // get view offset in cell space
    const offset = cell.project(plot.viewport, plot.zoom);

    // set blending func
    gl.enable(gl.BLEND);
    gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

    // bind shader
    shader.use();

    // set global uniforms
    shader.setUniform("uProjectionMatrix", proj);
    shader.setUniform("uViewOffset", [offset.x, offset.y]);
    shader.setUniform("uScale", scale);

    // for each quad buffer
    quads.forEach((buffer) => {
      // draw the points
      buffer.vertex.bind();
      buffer.vertex.draw();
    });
  }
  // renders IDs of the quads to a separate FBO
  renderIds() {
    const gl = this.gl;
    const quads = this.quads;
    const plot = this.overlay.plot;
    const cell = plot.cell;
    const proj = this.getOrthoMatrix();
    const scale = Math.pow(2, plot.zoom - cell.zoom);
    // get view offset in cell space
    const offset = cell.project(plot.viewport, plot.zoom);
    if (this.didCanvasResize(gl.canvas)) {
      // the canvas was resized, make the framebuffer attachments match
      this.setFramebufferAttachmentSizes(gl.canvas.width, gl.canvas.height);
    }
    this.pickingShader.use();
    gl.bindFramebuffer(gl.FRAMEBUFFER, this.fbo);
    gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
    gl.disable(gl.BLEND); // !important
    gl.enable(gl.DEPTH_TEST);

    // Clear the canvas AND the depth buffer.
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
    //uniforms
    this.pickingShader.setUniform("uProjectionMatrix", proj);
    this.pickingShader.setUniform("uViewOffset", [offset.x, offset.y]);
    this.pickingShader.setUniform("uScale", scale);
    quads.forEach((buffer) => {
      buffer.vertex.bind();
      buffer.vertex.draw();
    });
    gl.bindFramebuffer(gl.FRAMEBUFFER, null);
    gl.disable(gl.DEPTH_TEST);
  }

  /**
   * The draw function that is executed per frame.
   *
   * @returns {BatchquadOverlayRenderer} The overlay object, for chaining.
   */
  draw() {
    if (!this.quads) {
      return this;
    }
    this.renderColor(); // render color fbo (for users to see)
    this.renderIds(); // render IDS to fb (offscreen)
    return this;
  }
  // checks if the canvas has resized by checking the fboDimensions
  didCanvasResize(canvas) {
    return (
      canvas.width !== this.fboDimensions.width ||
      canvas.height !== this.fboDimensions.height
    );
  }
  // createFBO creates the ID FBO. Should only be called once.
  createFBO() {
    const gl = this.gl;
    this.targetTexture = gl.createTexture();
    gl.bindTexture(gl.TEXTURE_2D, this.targetTexture);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    // create a depth renderbuffer
    this.depthBuffer = gl.createRenderbuffer();
    gl.bindRenderbuffer(gl.RENDERBUFFER, this.depthBuffer);

    // Create and bind the framebuffer
    this.fbo = gl.createFramebuffer();
    gl.bindFramebuffer(gl.FRAMEBUFFER, this.fbo);

    // attach the texture as the first color attachment
    const attachmentPoint = gl.COLOR_ATTACHMENT0;
    const level = 0;
    gl.framebufferTexture2D(
      gl.FRAMEBUFFER,
      attachmentPoint,
      gl.TEXTURE_2D,
      this.targetTexture,
      level
    );

    // make a depth buffer and the same size as the targetTexture
    gl.framebufferRenderbuffer(
      gl.FRAMEBUFFER,
      gl.DEPTH_ATTACHMENT,
      gl.RENDERBUFFER,
      this.depthBuffer
    );
    gl.bindFramebuffer(gl.FRAMEBUFFER, null);
  }
  // readPixels reads the current bound FBO at the specified pixel location x,y
  readPixels(x, y) {
    const gl = this.gl;
    gl.flush();
    const data = new Uint8Array(4);
    gl.readPixels(
      x, // x
      y, // y
      1, // width
      1, // height
      gl.RGBA, // format
      gl.UNSIGNED_BYTE, // type
      data
    );
    const id = this.RGBAToId(data);
    return id - 1; // ids start at 1 -- 0 is reserved for background
  }
  // adds listeners to callback map -- please see EVENT_TYPES
  addListener(event, callback) {
    this.callbacks[event].push(callback);
  }
  // clears all listeners across all event types
  clearListeners() {
    const vals = Object.values(EVENT_TYPES);
    vals.forEach((val) => {
      this.callbacks[val] = [];
    });
  }
  // onClick read ID FBO at the pixel the mouse clicked on and extract the ID.
  onClick(event) {
    this.x = event.layerX;
    this.y = event.layerY;
    clearTimeout(this.hoverTimeoutId); // clear hover
    const gl = this.gl;
    gl.bindFramebuffer(gl.FRAMEBUFFER, this.fbo);
    // convert position to pixel space -- only applies if client width is different than the cavas size
    const pixelX = (this.x * gl.canvas.width) / gl.canvas.clientWidth;
    const pixelY =
      gl.canvas.height -
      (this.y * gl.canvas.height) / gl.canvas.clientHeight -
      1;
    const id = this.readPixels(pixelX, pixelY);
    // if id is the background id it means the user clicked on nothing -- do a clean up and dont invoke the listener's callbacks
    if (id === this.BACKGROUND_ID) {
      // clean up
      gl.bindFramebuffer(gl.FRAMEBUFFER, null);
      return;
    }
    this.callbacks[EVENT_TYPES.MOUSE_CLICK].forEach((cb) => {
      cb(id);
    });
  }
  // onMove register a callback that will invoke onHover if the mouse is not moved after the defined hoverThreshold time
  onMove(event) {
    this.x = event.layerX;
    this.y = event.layerY;
    //clear existing timeout, if mouse does not move for hoverThreshold time then we are hovering on something.
    clearTimeout(this.hoverTimeoutId);
    this.hoverTimeoutId = setTimeout(() => {
      this.onHover();
    }, this.hoverThreshold);
  }
  // onHover read ID FBO at the pixel the mouse is on and extract the ID. Then call all of the listeners for OnHover
  onHover() {
    const gl = this.gl;
    gl.bindFramebuffer(gl.FRAMEBUFFER, this.fbo);
    // convert position to pixel space -- only applies if client width is different than the cavas size
    const pixelX = (this.x * gl.canvas.width) / gl.canvas.clientWidth;
    const pixelY =
      gl.canvas.height -
      (this.y * gl.canvas.height) / gl.canvas.clientHeight -
      1;
    const id = this.readPixels(pixelX, pixelY);
    if (id === this.BACKGROUND_ID) {
      // clean up
      gl.bindFramebuffer(gl.FRAMEBUFFER, null);
      return;
    }
    this.callbacks[EVENT_TYPES.MOUSE_HOVER].forEach((cb) => {
      cb(id);
    });
    gl.bindFramebuffer(gl.FRAMEBUFFER, null);
  }
  // used to resize framebuffer when canvas has resized
  setFramebufferAttachmentSizes(width, height) {
    const gl = this.gl;
    gl.bindTexture(gl.TEXTURE_2D, this.targetTexture);
    // define size and format of level 0
    const level = 0;
    const internalFormat = gl.RGBA;
    const border = 0;
    const format = gl.RGBA;
    const type = gl.UNSIGNED_BYTE;
    const data = null;
    gl.texImage2D(
      gl.TEXTURE_2D,
      level,
      internalFormat,
      width,
      height,
      border,
      format,
      type,
      data
    );

    gl.bindRenderbuffer(gl.RENDERBUFFER, this.depthBuffer);
    gl.renderbufferStorage(
      gl.RENDERBUFFER,
      gl.DEPTH_COMPONENT16,
      width,
      height
    );
    this.fboDimensions.height = height;
    this.fboDimensions.width = width;
  }
  idToRGBA(id) {
    // 0 is reserved for background
    const ID = id + 1;
    return {
      iR: ((ID >> 0) & 0xff) / 0xff,
      iG: ((ID >> 8) & 0xff) / 0xff,
      iB: ((ID >> 16) & 0xff) / 0xff,
      iA: ((ID >> 24) & 0xff) / 0xff,
    };
  }
  RGBAToId(pixels) {
    return pixels[0] + (pixels[1] << 8) + (pixels[2] << 16) + (pixels[3] << 24);
  }
  /**
   *
   * @param {Array.<number>} latlng
   * @returns {{x:number, y:number}}
   * source https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
   */
  latlngToNormalized(latlng) {
    const maxLon = 180.0;
    const degreesToRadians = Math.PI / 180.0; // Factor for changing degrees to radians
    const latRadians = latlng[0] * degreesToRadians;
    const x = (latlng[1] + maxLon) / (maxLon * 2);
    const y =
      (1 -
        Math.log(Math.tan(latRadians) + 1 / Math.cos(latRadians)) / Math.PI) /
      2;

    return { x, y: 1 - y }; // have to invert y
  }
  /**
   *
   * @param {{x:number, y:number}} point
   * source https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
   */
  normalizedPointToLatLng(point) {
    const y = point.y; // invert y because the normalized functions inverts it
    const latRad = Math.atan(Math.sinh(Math.PI * (1 - 2 * y)));
    return { lat: -(latRad * (180 / Math.PI)), lng: point.x * 360 - 180 };
  }
}
