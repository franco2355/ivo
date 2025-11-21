import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import Spinner from './Spinner';

describe('Spinner Component', () => {
    it('should render with default props', () => {
        render(<Spinner />);
        const spinner = document.querySelector('.spinner');
        expect(spinner).toBeTruthy();
        expect(spinner.classList.contains('spinner-medium')).toBe(true);
    });

    it('should render with small size', () => {
        render(<Spinner size="small" />);
        const spinner = document.querySelector('.spinner');
        expect(spinner.classList.contains('spinner-small')).toBe(true);
    });

    it('should render with large size', () => {
        render(<Spinner size="large" />);
        const spinner = document.querySelector('.spinner');
        expect(spinner.classList.contains('spinner-large')).toBe(true);
    });

    it('should render with message', () => {
        const message = 'Cargando datos...';
        render(<Spinner message={message} />);
        expect(screen.getByText(message)).toBeTruthy();
    });

    it('should not render message when not provided', () => {
        render(<Spinner />);
        const message = document.querySelector('.spinner-message');
        expect(message).toBeFalsy();
    });
});
