const ejs = require('ejs');
const fs = require('fs');
const path = require('path');
const marked = require('marked');
const frontMatter = require('front-matter');
const ejsLayouts = require('ejs-layouts');

// Configuration
const config = {
    srcDir: 'site',
    distDir: 'docs',
    templatesDir: 'site/templates',
    contentDir: 'site/content',
    site: {
        name: 'Bold Brew',
        description: 'A modern TUI for Homebrew',
        url: 'https://bold-brew.com'
    }
};

// Function to generate a page
async function generatePage(template, data, outputPath) {
    const templatePath = path.join(config.templatesDir, template);
    const templateContent = fs.readFileSync(templatePath, 'utf-8');
    const layoutPath = path.join(config.templatesDir, 'layout.ejs');
    const layoutContent = fs.readFileSync(layoutPath, 'utf-8');
    
    // Render the template content
    const content = ejs.render(templateContent, {
        ...data,
        filename: templatePath
    });

    // Render the layout with the content
    const html = ejs.render(layoutContent, {
        ...data,
        filename: layoutPath,
        content
    });

    fs.mkdirSync(path.dirname(outputPath), { recursive: true });
    fs.writeFileSync(outputPath, html);
}

// Function to generate the homepage
async function generateHomepage() {
    const posts = getBlogPosts();
    await generatePage('index.ejs', {
        title: 'Bold Brew (bbrew) - Modern Homebrew TUI Manager for macOS and Linux',
        description: 'Bold Brew (bbrew) is the modern Terminal User Interface for Homebrew on macOS and Linux. Install, update, and manage packages and casks with an elegant TUI.',
        keywords: 'bbrew, Bold Brew, Homebrew TUI, macOS package manager, Linux package manager, Homebrew casks, Homebrew GUI, terminal package manager, Homebrew alternative, Project Bluefin, macOS development tools, Linux development tools',
        canonicalUrl: config.site.url,
        ogType: 'website',
        posts,
        site: config.site
    }, path.join(config.distDir, 'index.html'));
}

// Function to generate the blog
async function generateBlog() {
    // Generate the main blog page
    await generatePage('blog/index.ejs', {
        title: 'Blog | Bold Brew (bbrew)',
        description: 'Tips, tutorials, and guides for managing Homebrew packages on macOS',
        keywords: 'Homebrew blog, macOS tutorials, package management, Bold Brew guides',
        canonicalUrl: `${config.site.url}/blog/`,
        ogType: 'website',
        breadcrumb: [
            { text: 'Home', url: '/' },
            { text: 'Blog', url: '/blog/' }
        ],
        posts: getBlogPosts(),
        site: config.site
    }, path.join(config.distDir, 'blog/index.html'));

    // Generate article pages
    const blogDir = path.join(__dirname, config.contentDir, 'blog');
    if (fs.existsSync(blogDir)) {
        const files = fs.readdirSync(blogDir)
            .filter(file => file.endsWith('.md'));

        for (const file of files) {
            const filePath = path.join(blogDir, file);
            const content = fs.readFileSync(filePath, 'utf8');
            const { attributes, body } = frontMatter(content);
            const htmlContent = marked.parse(body);
            const outputFile = file.replace('.md', '.html');

            await generatePage('blog/post.ejs', {
                title: attributes.title || '',
                description: attributes.description || '',
                keywords: attributes.keywords || 'Homebrew, macOS, package management, Bold Brew, bbrew, terminal, development tools',
                date: attributes.date || '',
                content: htmlContent,
                canonicalUrl: `${config.site.url}/blog/${outputFile}`,
                ogType: 'article',
                breadcrumb: [
                    { text: 'Home', url: '/' },
                    { text: 'Blog', url: '/blog/' },
                    { text: attributes.title || '', url: `/blog/${outputFile}` }
                ],
                site: config.site
            }, path.join(config.distDir, 'blog', outputFile));
        }
    }
}

function getBlogPosts() {
    const blogDir = path.join(__dirname, config.contentDir, 'blog');
    const posts = [];
    
    if (!fs.existsSync(blogDir)) {
        return posts;
    }

    const files = fs.readdirSync(blogDir)
        .filter(file => file.endsWith('.md'));

    for (const file of files) {
        const content = fs.readFileSync(path.join(blogDir, file), 'utf8');
        const { attributes } = frontMatter(content);
        const outputFile = file.replace('.md', '.html');

        if (attributes.title && attributes.date) {
            posts.push({
                title: attributes.title,
                date: attributes.date,
                url: `/blog/${outputFile}`,
                excerpt: attributes.description || ''
            });
        }
    }

    return posts.sort((a, b) => new Date(b.date) - new Date(a.date));
}

// Function to generate the sitemap
async function generateSitemap() {
    const posts = getBlogPosts();
    const baseUrl = config.site.url;
    const today = new Date().toISOString().split('T')[0];

    // Static pages
    const staticPages = [
        {
            url: '/',
            lastmod: today,
            changefreq: 'weekly',
            priority: '1.0'
        },
        {
            url: '/blog/',
            lastmod: today,
            changefreq: 'weekly',
            priority: '0.9'
        }
    ];

    // Blog pages
    const blogPages = posts.map(post => ({
        url: post.url,
        lastmod: post.date,
        changefreq: 'monthly',
        priority: '0.8'
    }));

    // Combine all pages
    const allPages = [...staticPages, ...blogPages];

    // Generate XML content
    const sitemapContent = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${allPages.map(page => `  <url>
    <loc>${baseUrl}${page.url}</loc>
    <lastmod>${page.lastmod}</lastmod>
    ${page.changefreq ? `<changefreq>${page.changefreq}</changefreq>` : ''}
    ${page.priority ? `<priority>${page.priority}</priority>` : ''}
  </url>`).join('\n')}
</urlset>`;

    // Write the sitemap.xml file
    fs.writeFileSync(path.join(config.distDir, 'sitemap.xml'), sitemapContent);
}

// Main function
async function build() {
    try {
        // Clean the output directory while preserving assets, .git and other static files
        if (fs.existsSync(config.distDir)) {
            // Read all files in the docs directory
            const files = fs.readdirSync(config.distDir);
            
            // List of files/directories to preserve
            const preserveFiles = [
                'assets',
                '.git',
                'manifest.json',
                'robots.txt',
                'CNAME'
            ];
            
            // Remove only dynamically generated files
            for (const file of files) {
                if (!preserveFiles.includes(file)) {
                    const filePath = path.join(config.distDir, file);
                    // Check if it's a dynamically generated HTML file
                    if (file.endsWith('.html')) {
                        fs.rmSync(filePath, { recursive: true, force: true });
                    }
                }
            }
        } else {
            fs.mkdirSync(config.distDir);
        }
        
        // Generate pages
        await generateHomepage();
        await generateBlog();
        await generateSitemap();
        
        console.log('Build completed successfully!');
    } catch (error) {
        console.error('Build failed:', error);
        process.exit(1);
    }
}

build(); 