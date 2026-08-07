# Components V2 Guide

## Overview

Components V2 is Discord's updated message component system. It replaces
embeds and action rows with a richer layout model based on containers, sections,
media galleries, and separators. discord.go provides full V2 support through the
`components` package.

## V2 Message Model

A Components V2 message sets the `IS_COMPONENTS_V2` flag (bit 15) on the message
flags field. V2 messages use `components` as the primary content instead of
`embeds` or `content`:

```go
import (
    "github.com/discord-go/discord.go/components"
    "github.com/discord-go/discord.go/messages"
)

msg := messages.MessageSend{
    Flags: messages.FlagIsComponentsV2,
    Components: []components.Component{
        // V2 components go here
    },
}
```

When `FlagIsComponentsV2` is set:
- `content` is limited to 4000 characters.
- `embeds` is not allowed.
- Components can include V2-specific types: containers, sections, separators,
  and media galleries.

## Component Containers

`Container` groups components together with optional visual styling:

```go
container := components.Container{
    Components: []components.Component{
        textComponent,
        separator,
    },
    AccentColor: 0x5865F2,
    Spoiler:     false,
}
```

Containers can hold up to 10 components and support an accent color bar on the
left side.

## Media Galleries

`MediaGallery` displays up to 10 media items in a grid:

```go
gallery := components.MediaGallery{
    Items: []components.MediaGalleryItem{
        {URL: "https://example.com/image1.png"},
        {URL: "https://example.com/image2.png"},
    },
}
```

Each item supports a `Description` for accessibility and a `Spoiler` flag.

## Separators

`Separator` adds a visual divider between components:

```go
separator := components.Separator{
    Divider: true,
    Spacing: components.SeparatorSpacingSizeLarge,
}
```

Set `Divider` to `false` for a spacing-only separator without a line.

## Sections

`Section` groups text content with an accessory component (such as a button or
thumbnail):

```go
section := components.Section{
    Components: []components.Component{
        textComponent,
    },
    Accessory: buttonComponent,
}
```

The accessory appears on the right side of the text content.

## How V2 Differs From V1

| Aspect | V1 | V2 |
|--------|----|----|
| Layout | Action rows with buttons/select menus | Containers with flexible layout |
| Rich content | Embeds | Text components, media galleries |
| Separators | Not available | First-class separator component |
| Sections | Not available | Text with side accessory |
| Max depth | 1 (action row > component) | 3 (container > section > component) |

V1 components (action rows, buttons, select menus) still work in V2 messages
inside containers. V2 adds new layout primitives on top.

## Building V2 Messages

Use the `components` package builders to construct V2 messages:

```go
msg := messages.MessageSend{
    Flags: messages.FlagIsComponentsV2,
    Components: []components.Component{
        components.Container{
            Components: []components.Component{
                components.TextDisplay{Content: "# Welcome!"},
                components.Separator{Divider: true},
                components.MediaGallery{
                    Items: []components.MediaGalleryItem{
                        {URL: "https://example.com/banner.png", Description: "Server banner"},
                    },
                },
            },
            AccentColor: 0x5865F2,
        },
    },
}
```

## Common Patterns

- Use containers to group related components with a visual accent color.
- Use sections to pair text with a button or thumbnail.
- Use separators to create visual breaks between content blocks.
- Use media galleries for image grids.

## Best Practices

- Set `FlagIsComponentsV2` on the message flags.
- Do not use `embeds` in V2 messages; use text components instead.
- Keep container depth within Discord's limits (3 levels).
- Provide descriptions on media items for accessibility.

## Common Mistakes

- Setting `FlagIsComponentsV2` but still using `embeds`.
- Exceeding the 10-component limit per container.
- Forgetting that `content` is limited to 4000 characters in V2.
- Mixing V1 and V2 layout expectations.
